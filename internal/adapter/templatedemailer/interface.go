package templatedemailer

import (
	"log/slog"

	mg "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/emailer/mailgun"

	c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

// TemplatedEmailer Is adapter for responsive HTML email templates sender.
type TemplatedEmailer interface {
	SendNewUserTemporaryPasswordEmail(email, firstName, temporaryPassword string) error
	SendVerificationEmail(email, verificationCode, firstName string) error
	SendForgotPasswordEmail(email, verificationCode, firstName string) error
	GetDomainName() string
}

type templatedEmailer struct {
	UUID    uuid.Provider
	Logger  *slog.Logger
	Emailer mg.Emailer
}

func NewTemplatedEmailer(cfg *c.Conf, logger *slog.Logger, uuidp uuid.Provider, emailer mg.Emailer) TemplatedEmailer {
	// Defensive code: Make sure we have access to the file before proceeding any further with the code.
	logger.Debug("templated emailer initializing...")
	logger.Debug("templated emailer initialized")

	return &templatedEmailer{
		UUID:    uuidp,
		Logger:  logger,
		Emailer: emailer,
	}
}
func (impl *templatedEmailer) GetDomainName() string {
	return impl.Emailer.GetDomainName()
}
