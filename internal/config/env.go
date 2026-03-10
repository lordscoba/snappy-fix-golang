package config

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Server         ServerConfiguration
	Database       Database
	TestDatabase   Database
	App            App
	IPStack        IPStack
	IPData         IPData
	IPAPI          IPAPI
	Mail           MAIL
	Redis          Redis
	Mailgun        Mailgun
	AmazonS3Bucket AmazonS3Bucket
	Paystack       Paystack
	Cloudinary     Cloudinary
}

type BaseConfig struct {
	SERVER_PORT                       string  `mapstructure:"SERVER_PORT"`
	SERVER_SECRET                     string  `mapstructure:"SERVER_SECRET"`
	SERVER_ACCESSTOKENEXPIREDURATION  int     `mapstructure:"SERVER_ACCESSTOKENEXPIREDURATION"`
	SERVER_REFRESHTOKENEXPIREDURATION int     `mapstructure:"SERVER_REFRESHTOKENEXPIREDURATION"`
	REQUEST_PER_SECOND                float64 `mapstructure:"REQUEST_PER_SECOND"`
	TRUSTED_PROXIES                   string  `mapstructure:"TRUSTED_PROXIES"`
	EXEMPT_FROM_THROTTLE              string  `mapstructure:"EXEMPT_FROM_THROTTLE"`

	APP_NAME                string `mapstructure:"APP_NAME"`
	APP_MODE                string `mapstructure:"APP_MODE"`
	APP_URL                 string `mapstructure:"APP_URL"`
	RESET_PASSWORD_DURATION int    `mapstructure:"RESET_PASSWORD_DURATION"`
	APP_ENV                 string `mapstructure:"APP_ENV"`
	APP_FRONTEND_BASE_URL   string `mapstructure:"APP_FRONTEND_BASE_URL"`

	DB_HOST  string `mapstructure:"DB_HOST"`
	DB_PORT  string `mapstructure:"DB_PORT"`
	TIMEZONE string `mapstructure:"TIMEZONE"`
	SSLMODE  string `mapstructure:"SSLMODE"`
	USERNAME string `mapstructure:"USERNAME"`
	PASSWORD string `mapstructure:"PASSWORD"`
	DB_NAME  string `mapstructure:"DB_NAME"`
	MIGRATE  bool   `mapstructure:"MIGRATE"`

	TEST_DB_HOST  string `mapstructure:"TEST_DB_HOST"`
	TEST_DB_PORT  string `mapstructure:"TEST_DB_PORT"`
	TEST_TIMEZONE string `mapstructure:"TEST_TIMEZONE"`
	TEST_SSLMODE  string `mapstructure:"TEST_SSLMODE"`
	TEST_USERNAME string `mapstructure:"TEST_USERNAME"`
	TEST_PASSWORD string `mapstructure:"TEST_PASSWORD"`
	TEST_DB_NAME  string `mapstructure:"TEST_DB_NAME"`
	TEST_MIGRATE  bool   `mapstructure:"TEST_MIGRATE"`

	IPSTACK_KEY      string `mapstructure:"IPSTACK_KEY"`
	IPSTACK_BASE_URL string `mapstructure:"IPSTACK_BASE_URL"`

	IPDATA_KEY      string `mapstructure:"IPDATA_KEY"`
	IPDATA_BASE_URL string `mapstructure:"IPDATA_BASE_URL"`

	IPAPI_BASE_URL string `mapstructure:"IPAPI_BASE_URL"`

	MAIL_SERVER   string `mapstructure:"MAIL_SERVER"`
	MAIL_PASSWORD string `mapstructure:"MAIL_PASSWORD"`
	MAIL_USERNAME string `mapstructure:"MAIL_USERNAME"`
	MAIL_PORT     string `mapstructure:"MAIL_PORT"`

	REDIS_PORT string `mapstructure:"REDIS_PORT"`
	REDIS_HOST string `mapstructure:"REDIS_HOST"`
	REDIS_DB   string `mapstructure:"REDIS_DB"`

	MAILGUN_BASE_URL string `mapstructure:"MAILGUN_BASE_URL"`
	MAILGUN_API_KEY  string `mapstructure:"MAILGUN_API_KEY"`
	MAILGUN_DOMAIN   string `mapstructure:"MAILGUN_DOMAIN"`

	FCM_BASE_URL     string `mapstructure:"FCM_BASE_URL"`
	FCM_PROJECT_ID   string `mapstructure:"FCM_PROJECT_ID"`
	FCM_OAUTH_BEARER string `mapstructure:"FCM_OAUTH_BEARER"`

	AWS_ACCESS_KEY_ID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWS_SECRET_ACCESS_KEY string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWS_REGION            string `mapstructure:"AWS_REGION"`
	AWS_S3_BUCKET         string `mapstructure:"AWS_S3_BUCKET"`
	AWS_S3_PUBLIC_BASE    string `mapstructure:"AWS_S3_PUBLIC_BASE"`

	E2EE_ENABLED                string `mapstructure:"E2EE_ENABLED"`
	E2EE_CONFIG_JSON            string `mapstructure:"E2EE_CONFIG_JSON"`
	E2EE_SERVER_PRIVATE_JWK_B64 string `mapstructure:"E2EE_SERVER_PRIVATE_JWK_B64"`
	E2EE_SERVER_PRIVATE_JWK_ENC string `mapstructure:"E2EE_SERVER_PRIVATE_JWK_ENC"`
	E2EE_AT_REST_PASSPHRASE     string `mapstructure:"E2EE_AT_REST_PASSPHRASE"`
	E2EE_AT_REST_RAWKEY_B64     string `mapstructure:"E2EE_AT_REST_RAWKEY_B64"`

	PAYSTACK_BASE_URL   string `mapstructure:"PAYSTACK_BASE_URL"`
	PAYSTACK_SECRET_KEY string `mapstructure:"PAYSTACK_SECRET_KEY"`
	PAYSTACK_PUBLIC_KEY string `mapstructure:"PAYSTACK_PUBLIC_KEY"`

	TWILIO_BASE_URL        string `mapstructure:"TWILIO_BASE_URL"`
	TWILIO_ACCOUNT_SID     string `mapstructure:"TWILIO_ACCOUNT_SID"`
	TWILIO_AUTH_TOKEN      string `mapstructure:"TWILIO_AUTH_TOKEN"`
	TWILIO_PHONE_NUMBER    string `mapstructure:"TWILIO_PHONE_NUMBER"`
	TWILIO_WHATSAPP_NUMBER string `mapstructure:"TWILIO_WHATSAPP_NUMBER"`

	TWILIO_VERIFY_BASE_URL           string `mapstructure:"TWILIO_VERIFY_BASE_URL"`
	TWILIO_VERIFY_SERVICE_SID        string `mapstructure:"TWILIO_VERIFY_SERVICE_SID"`
	TWILIO_NOTIFY_BASE_URL           string `mapstructure:"TWILIO_NOTIFY_BASE_URL"`
	TWILIO_NOTIFY_SERVICE_SID        string `mapstructure:"TWILIO_NOTIFY_SERVICE_SID"`
	TWILIO_API_KEY_SID               string `mapstructure:"TWILIO_API_KEY_SID"`
	TWILIO_API_KEY_SECRET            string `mapstructure:"TWILIO_API_KEY_SECRET"`
	TWILIO_CONVERSATIONS_SERVICE_SID string `mapstructure:"TWILIO_CONVERSATIONS_SERVICE_SID"`

	KORAPAY_BASE_URL   string `mapstructure:"KORAPAY_BASE_URL"`
	KORAPAY_SECRET_KEY string `mapstructure:"KORAPAY_SECRET_KEY"`
	KORAPAY_PUBLIC_KEY string `mapstructure:"KORAPAY_PUBLIC_KEY"`

	SENDCHAMP_BASE_URL string `mapstructure:"SENDCHAMP_BASE_URL"`
	SENDCHAMP_API_KEY  string `mapstructure:"SENDCHAMP_API_KEY"`

	ESMSAFRICA_BASE_URL   string `mapstructure:"ESMSAFRICA_BASE_URL"`
	ESMSAFRICA_ACCOUNT_ID string `mapstructure:"ESMSAFRICA_ACCOUNT_ID"`
	ESMSAFRICA_API_KEY    string `mapstructure:"ESMSAFRICA_API_KEY"`

	CLOUDINARY_CLOUD_NAME string `mapstructure:"CLOUDINARY_CLOUD_NAME"`
	CLOUDINARY_API_KEY    string `mapstructure:"CLOUDINARY_API_KEY"`
	CLOUDINARY_API_SECRET string `mapstructure:"CLOUDINARY_API_SECRET"`
}

func (config *BaseConfig) SetupConfigurationn() *Configuration {
	trustedProxies := []string{}
	exemptFromThrottle := []string{}
	json.Unmarshal([]byte(config.TRUSTED_PROXIES), &trustedProxies)
	json.Unmarshal([]byte(config.EXEMPT_FROM_THROTTLE), &exemptFromThrottle)
	if config.SERVER_PORT == "" {
		config.SERVER_PORT = os.Getenv("PORT")
	}
	return &Configuration{
		Server: ServerConfiguration{
			Port:                       config.SERVER_PORT,
			Secret:                     config.SERVER_SECRET,
			AccessTokenExpireDuration:  config.SERVER_ACCESSTOKENEXPIREDURATION,
			RefreshTokenExpireDuration: config.SERVER_REFRESHTOKENEXPIREDURATION,
			RequestPerSecond:           config.REQUEST_PER_SECOND,
			TrustedProxies:             trustedProxies,
			ExemptFromThrottle:         exemptFromThrottle,
		},
		App: App{
			Name:                  config.APP_NAME,
			Mode:                  config.APP_MODE,
			Url:                   config.APP_URL,
			AppEnv:                config.APP_ENV,
			ResetPasswordDuration: config.RESET_PASSWORD_DURATION,
			FrontendUrl:           config.APP_FRONTEND_BASE_URL,
		},
		Database: Database{
			DB_HOST:  config.DB_HOST,
			DB_PORT:  config.DB_PORT,
			USERNAME: config.USERNAME,
			PASSWORD: config.PASSWORD,
			TIMEZONE: config.TIMEZONE,
			SSLMODE:  config.SSLMODE,
			DB_NAME:  config.DB_NAME,
			Migrate:  config.MIGRATE,
		},
		TestDatabase: Database{
			DB_HOST:  config.TEST_DB_HOST,
			DB_PORT:  config.TEST_DB_PORT,
			USERNAME: config.TEST_USERNAME,
			PASSWORD: config.TEST_PASSWORD,
			TIMEZONE: config.TEST_TIMEZONE,
			SSLMODE:  config.TEST_SSLMODE,
			DB_NAME:  config.TEST_DB_NAME,
			Migrate:  config.TEST_MIGRATE,
		},

		IPStack: IPStack{
			Key:     config.IPSTACK_KEY,
			BaseUrl: config.IPSTACK_BASE_URL,
		},

		IPData: IPData{
			Key:     config.IPDATA_KEY,
			BaseUrl: config.IPDATA_BASE_URL,
		},

		IPAPI: IPAPI{
			BaseUrl: config.IPAPI_BASE_URL,
		},

		Mail: MAIL{
			Server:   config.MAIL_SERVER,
			Password: config.MAIL_PASSWORD,
			Port:     config.MAIL_PORT,
			Username: config.MAIL_USERNAME,
		},

		Redis: Redis{
			REDIS_PORT: config.REDIS_PORT,
			REDIS_HOST: config.REDIS_HOST,
			REDIS_DB:   config.REDIS_DB,
		},

		Mailgun: Mailgun{
			BaseUrl: config.MAILGUN_BASE_URL,
			Domain:  config.MAILGUN_DOMAIN,
			ApiKey:  config.MAILGUN_API_KEY,
		},

		AmazonS3Bucket: AmazonS3Bucket{
			AccessKeyID:     config.AWS_ACCESS_KEY_ID,
			SecretAccessKey: config.AWS_SECRET_ACCESS_KEY,
			Region:          config.AWS_REGION,
			Bucket:          config.AWS_S3_BUCKET,
			PublicBase:      config.AWS_S3_PUBLIC_BASE,
		},

		Paystack: Paystack{
			PaystackSecretKey: config.PAYSTACK_SECRET_KEY,
			PaystackPublicKey: config.PAYSTACK_PUBLIC_KEY,
			PaystackBaseUrl:   config.PAYSTACK_BASE_URL,
		},

		Cloudinary: Cloudinary{
			CloudName: config.CLOUDINARY_CLOUD_NAME,
			APIKey:    config.CLOUDINARY_API_KEY,
			APISecret: config.CLOUDINARY_API_SECRET,
		},
	}
}
