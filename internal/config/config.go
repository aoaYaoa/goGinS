package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// 服务
	ServerPort  string
	ServerMode  string
	FrontendURL string
	CORSOrigins []string

	// 数据库（PostgreSQL）
	DatabaseHost      string
	DatabasePort      string
	DatabaseName      string
	DatabaseUser      string
	DatabasePass      string
	DatabaseSSLMode   string
	MigrationsPath    string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime int // 秒
	DBConnMaxIdleTime int // 秒

	// JWT
	JWTSecret string

	// 签名与加密
	EnableSignature  bool
	SignatureSecret  string
	EncryptionAESKey string

	// Redis
	RedisAddr     string
	RedisUsername string
	RedisPassword string
	RedisDB       int
	RedisTLS      bool

	// Kafka
	KafkaBrokers          string
	KafkaTopic            string
	KafkaSecurityProtocol string
	KafkaSSLCAFile        string
	KafkaSSLCertFile      string
	KafkaSSLKeyFile       string

	// IP 访问控制
	EnableIPWhitelist bool
	IPWhitelist       string
	EnableIPBlacklist bool
	IPBlacklist       string

	// 可观测性
	MetricsAllowedIPs string
	OTELEndpoint      string
	OTELServiceName   string

	// 文件存储
	StorageType          string
	LocalStorageDir      string
	StoragePublicBaseURL string
	S3Bucket             string
	S3Region             string
	S3Endpoint           string
	S3AccessKey          string
	S3SecretKey          string
	S3UsePathStyle       bool

	// 通知
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

var AppConfig Config

func Init() {
	_ = godotenv.Load()

	port := getEnv("SERVER_PORT", "")
	if port == "" {
		port = getEnv("PORT", "8080")
	}

	AppConfig = Config{
		ServerPort:  port,
		ServerMode:  getEnv("SERVER_MODE", "debug"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		CORSOrigins: splitCSV(getEnv("CORS_ORIGIN", "http://localhost:5173")),

		DatabaseHost:      getEnv("DATABASE_HOST", "localhost"),
		DatabasePort:      getEnv("DATABASE_PORT", "5432"),
		DatabaseName:      getEnv("DATABASE_NAME", "appdb"),
		DatabaseUser:      getEnv("DATABASE_USER", "postgres"),
		DatabasePass:      getEnv("DATABASE_PASS", ""),
		DatabaseSSLMode:   getEnv("DATABASE_SSL_MODE", "disable"),
		MigrationsPath:    getEnv("MIGRATIONS_PATH", "migrations"),
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
		DBConnMaxIdleTime: getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 60),

		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),

		EnableSignature:  getEnvAsBool("ENABLE_SIGNATURE", false),
		SignatureSecret:  getEnv("SIGNATURE_SECRET", ""),
		EncryptionAESKey: getEnv("ENCRYPTION_AES_KEY", ""),

		RedisAddr:     getEnv("REDIS_ADDR", ""),
		RedisUsername: getEnv("REDIS_USERNAME", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		RedisTLS:      getEnvAsBool("REDIS_TLS", false),

		KafkaBrokers:          getEnv("KAFKA_BROKERS", ""),
		KafkaTopic:            getEnv("KAFKA_TOPIC", ""),
		KafkaSecurityProtocol: getEnv("KAFKA_SECURITY_PROTOCOL", ""),
		KafkaSSLCAFile:        getEnv("KAFKA_SSL_CA_FILE", ""),
		KafkaSSLCertFile:      getEnv("KAFKA_SSL_CERT_FILE", ""),
		KafkaSSLKeyFile:       getEnv("KAFKA_SSL_KEY_FILE", ""),

		EnableIPWhitelist: getEnvAsBool("ENABLE_IP_WHITELIST", false),
		IPWhitelist:       getEnv("IP_WHITELIST", ""),
		EnableIPBlacklist: getEnvAsBool("ENABLE_IP_BLACKLIST", false),
		IPBlacklist:       getEnv("IP_BLACKLIST", ""),

		MetricsAllowedIPs: getEnv("METRICS_ALLOWED_IPS", ""),
		OTELEndpoint:      getEnv("OTEL_ENDPOINT", ""),
		OTELServiceName:   getEnv("OTEL_SERVICE_NAME", "go-gin-starter"),

		StorageType:          getEnv("STORAGE_TYPE", "local"),
		LocalStorageDir:      getEnv("LOCAL_STORAGE_DIR", "uploads"),
		StoragePublicBaseURL: getEnv("STORAGE_PUBLIC_BASE_URL", "/uploads"),
		S3Bucket:             getEnv("S3_BUCKET", ""),
		S3Region:             getEnv("S3_REGION", ""),
		S3Endpoint:           getEnv("S3_ENDPOINT", ""),
		S3AccessKey:          getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:          getEnv("S3_SECRET_KEY", ""),
		S3UsePathStyle:       getEnvAsBool("S3_USE_PATH_STYLE", true),

		SMTPHost: getEnv("SMTP_HOST", ""),
		SMTPPort: getEnv("SMTP_PORT", "587"),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", ""),
	}
}

func GetIPWhitelist() []string {
	if !AppConfig.EnableIPWhitelist || AppConfig.IPWhitelist == "" {
		return []string{}
	}
	return splitCSV(AppConfig.IPWhitelist)
}

func GetIPBlacklist() []string {
	if !AppConfig.EnableIPBlacklist || AppConfig.IPBlacklist == "" {
		return []string{}
	}
	return splitCSV(AppConfig.IPBlacklist)
}

func GetMetricsAllowedIPs() []string {
	if AppConfig.MetricsAllowedIPs == "" {
		return []string{}
	}
	return splitCSV(AppConfig.MetricsAllowedIPs)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func splitCSV(v string) []string {
	items := strings.Split(v, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
