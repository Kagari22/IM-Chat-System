package app

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	goredis "github.com/redis/go-redis/v9"

	"IM_Chat_System/internal/chat"
	"IM_Chat_System/internal/config"
	"IM_Chat_System/internal/handler"
	"IM_Chat_System/internal/model"
	"IM_Chat_System/internal/mq"
	redisPresence "IM_Chat_System/internal/presence/redis"
	"IM_Chat_System/internal/ratelimit"
	redisRateLimit "IM_Chat_System/internal/ratelimit/redis"
	mysqlrepo "IM_Chat_System/internal/repository/mysql"
	"IM_Chat_System/internal/search"
	essearch "IM_Chat_System/internal/search/elasticsearch"
	"IM_Chat_System/internal/service"
	"IM_Chat_System/internal/storage"
	miniostorage "IM_Chat_System/internal/storage/minio"
	"IM_Chat_System/internal/tokenblacklist"
	redisBlacklist "IM_Chat_System/internal/tokenblacklist/redis"
	redisUnread "IM_Chat_System/internal/unread/redis"
)

type App struct {
	Config    config.Config
	db        *sql.DB
	redis     *goredis.Client
	auth      *handler.AuthHandler
	logout    *handler.LogoutHandler
	users     *handler.UserHandler
	messages  *handler.MessageHandler
	media     *handler.MediaHandler
	search    *handler.SearchHandler
	hub       *chat.Hub
	blacklist tokenblacklist.Store
	limiter   ratelimit.Store
	publisher mq.EventPublisher
	consumer  *mq.MessageCreatedConsumer
	cancel    context.CancelFunc
	indexer   search.Indexer
}

func New() (*App, error) {
	cfg := config.Load()

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	userRepo := mysqlrepo.NewUserRepository(db)
	messageRepo := mysqlrepo.NewMessageRepository(db)
	presenceStore := redisPresence.New(rdb)
	unreadStore := redisUnread.New(rdb)
	blacklistStore := redisBlacklist.New(rdb)
	rateLimitStore := redisRateLimit.New(rdb)
	var uploader storage.Uploader = storage.NoopUploader{}
	if cfg.EnableMinIO {
		minioUploader, err := miniostorage.New(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL, cfg.MinIOPublicBaseURL)
		if err != nil {
			return nil, err
		}
		uploader = minioUploader
	}

	var indexer search.Indexer = search.NoopIndexer{}
	if cfg.EnableElasticsearch {
		esIndexer, err := essearch.New(cfg.ElasticsearchURL, cfg.ElasticsearchIndex)
		if err != nil {
			return nil, err
		}
		indexer = esIndexer
	}

	var publisher mq.EventPublisher = mq.NoopPublisher{}
	var consumer *mq.MessageCreatedConsumer
	var cancel context.CancelFunc

	if cfg.EnableRabbitMQ {
		rabbitPublisher, err := mq.NewRabbitPublisher(cfg.RabbitMQURL)
		if err != nil {
			return nil, err
		}
		publisher = rabbitPublisher

		consumer, err = mq.NewMessageCreatedConsumer(cfg.RabbitMQURL)
		if err != nil {
			_ = publisher.Close()
			return nil, err
		}
	}

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, int64(cfg.TokenTTL.Hours()))
	messageService := service.NewMessageService(userRepo, messageRepo, unreadStore, publisher, uploader)
	hub := chat.NewHub(messageService, cfg.JWTSecret, cfg.NodeID, presenceStore, blacklistStore, rateLimitStore)

	if consumer != nil {
		consumerCtx, consumerCancel := context.WithCancel(context.Background())
		cancel = consumerCancel
		if err := consumer.Start(consumerCtx, func(ctx context.Context, event mq.MessageCreatedEvent) error {
			if err := hub.DispatchMessageCreated(ctx, event); err != nil {
				return err
			}
			return indexer.IndexMessage(ctx, model.Message{
				ID:          event.MessageID,
				FromUserID:  event.FromUserID,
				ToUserID:    event.ToUserID,
				ContentType: event.ContentType,
				Content:     event.Content,
				ObjectKey:   event.ObjectKey,
				ObjectURL:   event.ObjectURL,
				FileName:    event.FileName,
				FileSize:    event.FileSize,
				CreatedAt:   event.CreatedAt,
			})
		}); err != nil {
			_ = consumer.Close()
			_ = publisher.Close()
			return nil, err
		}
	}

	return &App{
		Config:    cfg,
		db:        db,
		redis:     rdb,
		auth:      handler.NewAuthHandler(authService),
		logout:    handler.NewLogoutHandler(blacklistStore),
		users:     handler.NewUserHandler(messageService),
		messages:  handler.NewMessageHandler(messageService),
		media:     handler.NewMediaHandler(messageService),
		search:    handler.NewSearchHandler(indexer),
		hub:       hub,
		blacklist: blacklistStore,
		limiter:   rateLimitStore,
		publisher: publisher,
		consumer:  consumer,
		cancel:    cancel,
		indexer:   indexer,
	}, nil
}

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/register", handler.WithRateLimit(a.limiter, "register", 10, time.Minute, handler.ClientIPKey, a.auth.Register))
	mux.HandleFunc("POST /api/login", handler.WithRateLimit(a.limiter, "login", 20, time.Minute, handler.ClientIPKey, a.auth.Login))
	mux.HandleFunc("POST /api/logout", handler.WithAuth(a.Config.JWTSecret, a.blacklist, a.logout.Logout))
	mux.HandleFunc("GET /api/me", handler.WithAuth(a.Config.JWTSecret, a.blacklist, a.users.Me))
	mux.HandleFunc("GET /api/users", handler.WithAuth(a.Config.JWTSecret, a.blacklist, a.users.Users))
	mux.HandleFunc("GET /api/messages", handler.WithAuth(a.Config.JWTSecret, a.blacklist, a.messages.Messages))
	mux.HandleFunc("GET /api/offline", handler.WithAuth(a.Config.JWTSecret, a.blacklist, a.messages.Offline))
	mux.HandleFunc("POST /api/media/upload", handler.WithAuth(a.Config.JWTSecret, a.blacklist, a.media.Upload))
	mux.HandleFunc("GET /api/search/messages", handler.WithAuth(a.Config.JWTSecret, a.blacklist, a.search.SearchMessages))
	mux.HandleFunc("GET /ws", a.hub.ServeWS)
	mux.Handle("/", http.FileServer(http.Dir("web")))
	return logRequest(mux)
}

func (a *App) Close() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.consumer != nil {
		_ = a.consumer.Close()
	}
	if a.publisher != nil {
		_ = a.publisher.Close()
	}
	if a.indexer != nil {
		_ = a.indexer.Close()
	}
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
