package config

import (
	"log"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conf struct {
	InitialAccount initialAccountConf
	AppServer      serverConf
	DB             dbConfig
	AWS            awsConfig
	Emailer        mailgunConfig
	PDFBuilder     pdfBuilderConfig
}

type initialAccountConf struct {
	AdminEmail              string
	AdminPassword           string
	AdminTenantID           primitive.ObjectID
	AdminTenantName         string
	AdminTenantOpenAIAPIKey string
	AdminTenantOpenAIOrgKey string
}

type serverConf struct {
	Port                    string
	IP                      string
	HMACSecret              []byte
	HasDebugging            bool
	DomainName              string
	Enable2FAOnRegistration bool
}

type dbConfig struct {
	URI  string
	Name string
}

type awsConfig struct {
	AccessKey      string
	SecretKey      string
	Endpoint       string
	Region         string
	BucketName     string
	SSECustomerKey string
}

type mailgunConfig struct {
	APIKey      string
	Domain      string
	APIBase     string
	SenderEmail string
}

type pdfBuilderConfig struct {
	AssociateInvoiceTemplatePath string
	DataDirectoryPath            string
}

func New() *Conf {
	var c Conf
	c.InitialAccount.AdminEmail = getEnv("NONPROFITVAULT_BACKEND_INITIAL_ADMIN_EMAIL", true)
	c.InitialAccount.AdminPassword = getEnv("NONPROFITVAULT_BACKEND_INITIAL_ADMIN_PASSWORD", true)
	c.InitialAccount.AdminTenantID = getObjectIDEnv("NONPROFITVAULT_BACKEND_INITIAL_ADMIN_STORE_ID", true)
	c.InitialAccount.AdminTenantName = getEnv("NONPROFITVAULT_BACKEND_INITIAL_ADMIN_STORE_NAME", true)
	c.InitialAccount.AdminTenantOpenAIAPIKey = getEnv("NONPROFITVAULT_BACKEND_INITIAL_ADMIN_STORE_OPENAI_KEY", true)
	c.InitialAccount.AdminTenantOpenAIOrgKey = getEnv("NONPROFITVAULT_BACKEND_INITIAL_ADMIN_STORE_OPENAI_ORGANIZATION_KEY", true)

	c.AppServer.Port = getEnv("NONPROFITVAULT_BACKEND_PORT", true)
	c.AppServer.IP = getEnv("NONPROFITVAULT_BACKEND_IP", false)
	c.AppServer.HMACSecret = []byte(getEnv("NONPROFITVAULT_BACKEND_HMAC_SECRET", true))
	c.AppServer.HasDebugging = getEnvBool("NONPROFITVAULT_BACKEND_HAS_DEBUGGING", true, true)
	c.AppServer.DomainName = getEnv("NONPROFITVAULT_BACKEND_DOMAIN_NAME", true)
	c.AppServer.Enable2FAOnRegistration = getEnvBool("NONPROFITVAULT_BACKEND_APP_ENABLE_2FA_ON_REGISTRATION", false, false)

	c.DB.URI = getEnv("NONPROFITVAULT_BACKEND_DB_URI", true)
	c.DB.Name = getEnv("NONPROFITVAULT_BACKEND_DB_NAME", true)

	c.AWS.AccessKey = getEnv("NONPROFITVAULT_BACKEND_AWS_ACCESS_KEY", true)
	c.AWS.SecretKey = getEnv("NONPROFITVAULT_BACKEND_AWS_SECRET_KEY", true)
	c.AWS.Endpoint = getEnv("NONPROFITVAULT_BACKEND_AWS_ENDPOINT", true)
	c.AWS.Region = getEnv("NONPROFITVAULT_BACKEND_AWS_REGION", true)
	c.AWS.BucketName = getEnv("NONPROFITVAULT_BACKEND_AWS_BUCKET_NAME", true)
	// c.AWS.SSECustomerKey = getEnv("NONPROFITVAULT_BACKEND_AWS_SSE_CUSTOMER_KEY", false) // Experimental: Do not use yet.

	c.Emailer.APIKey = getEnv("NONPROFITVAULT_BACKEND_MAILGUN_API_KEY", true)
	c.Emailer.Domain = getEnv("NONPROFITVAULT_BACKEND_MAILGUN_DOMAIN", true)
	c.Emailer.APIBase = getEnv("NONPROFITVAULT_BACKEND_MAILGUN_API_BASE", true)
	c.Emailer.SenderEmail = getEnv("NONPROFITVAULT_BACKEND_MAILGUN_SENDER_EMAIL", true)

	c.PDFBuilder.DataDirectoryPath = getEnv("NONPROFITVAULT_BACKEND_PDF_BUILDER_DATA_DIRECTORY_PATH", true)
	c.PDFBuilder.AssociateInvoiceTemplatePath = getEnv("NONPROFITVAULT_BACKEND_PDF_BUILDER_ASSOCIATE_INVOICE_PATH", true)

	return &c
}

func getEnv(key string, required bool) string {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return value
}

func getEnvBool(key string, required bool, defaultValue bool) bool {
	valueStr := getEnv(key, required)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Fatalf("Invalid boolean value for environment variable %s", key)
	}
	return value
}

func getObjectIDEnv(key string, required bool) primitive.ObjectID {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	objectID, err := primitive.ObjectIDFromHex(value)
	if err != nil {
		log.Fatalf("Invalid mongodb primitive object id value for environment variable %s", key)
	}
	return objectID
}

func getByteArrayEnv(key string, required bool) []byte {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return []byte(value)
}
