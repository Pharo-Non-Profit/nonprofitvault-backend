package controller

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

type OTPGenerateResponseIDO struct {
	Base32     string `json:"base32"`
	OTPAuthURL string `json:"otpauth_url"`
}

// GenerateOTP function generates the time-based one-time password (TOTP) secret for the user. The user must use these values to generate a QR to present to the user.
func (impl *GatewayControllerImpl) GenerateOTP(ctx context.Context) (*OTPGenerateResponseIDO, error) {
	// Extract from our session the following data.
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	sessionID, _ := ctx.Value(constants.SessionID).(string)

	////
	//// Start the transaction.
	////

	session, err := impl.DbClient.StartSession()
	if err != nil {
		impl.Logger.Error("start session error",
			slog.Any("error", err))
		return nil, err
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {

		// Lookup the user in our database, else return a `400 Bad Request` error.
		u, err := impl.UserStorer.GetByID(sessCtx, userID)
		if err != nil {
			impl.Logger.Error("failed getting session user", slog.Any("err", err))
			return nil, err
		}
		if u == nil {
			impl.Logger.Warn("user does not exist validation error")
			return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
		}

		res := &OTPGenerateResponseIDO{}

		// Only generate a new OTP if no previous secret was generated. If
		// previously generated then reuse existing opt secret and otp auth url.
		if u.OTPSecret == "" && u.OTPAuthURL == "" {
			// STEP 1: Generate the OPT.
			key, err := totp.Generate(totp.GenerateOpts{
				Issuer:      impl.TemplatedEmailer.GetDomainName(),
				AccountName: u.Email,
				SecretSize:  15,
			})
			if err != nil {
				impl.Logger.Error("failed generating otp", slog.Any("err", err))
				return nil, err
			}

			// STEP 2: Save the secret to the user's profile.
			u.OTPSecret = key.Secret()
			u.OTPAuthURL = key.URL()
			u.ModifiedAt = time.Now()

			if err := impl.UserStorer.UpdateByID(sessCtx, u); err != nil {
				impl.Logger.Error("failed updating session user with opt secret", slog.Any("err", err))
				return nil, err
			}

			// STEP 3: Update the authenticated user session.
			uBin, err := json.Marshal(u)
			if err != nil {
				impl.Logger.Error("marshalling error", slog.Any("err", err))
				return nil, err
			}
			atExpiry := 14 * 24 * time.Hour
			err = impl.Cache.SetWithExpiry(sessCtx, sessionID, uBin, atExpiry)
			if err != nil {
				impl.Logger.Error("cache set with expiry error", slog.Any("err", err))
				return nil, err
			}

			// STEP 4: Share the secret with the user so they may give to their
			// third-party authenticator app.
			res.Base32 = key.Secret()
			res.OTPAuthURL = key.URL()

			impl.Logger.Debug("successfully generated opt secret and auth url", slog.Any("base_32", res.Base32), slog.Any("opt_auth_url", res.OTPAuthURL))
		} else {
			// Reuse the existing opt secret and auth url.
			res.Base32 = u.OTPSecret
			res.OTPAuthURL = u.OTPAuthURL
			impl.Logger.Warn("reusing previously generated opt secret and auth url", slog.Any("base_32", res.Base32), slog.Any("opt_auth_url", res.OTPAuthURL))
		}

		return res, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*OTPGenerateResponseIDO), nil
}

func (impl *GatewayControllerImpl) GenerateOTPAndQRCodePNGImage(ctx context.Context) ([]byte, error) {
	otpResponse, err := impl.GenerateOTP(ctx)
	if err != nil {
		impl.Logger.Error("failed generating otp",
			slog.Any("error", err))
		return nil, err
	}

	// Generate the QR code for the specific URL and return the `png` binary
	// file bytes.
	var png []byte
	png, err = qrcode.Encode(otpResponse.OTPAuthURL, qrcode.Medium, 256)
	if err != nil {
		impl.Logger.Error("encode error", slog.Any("error", err))
		return nil, err
	}

	impl.Logger.Debug("qr code ready",
		slog.Any("payload", otpResponse.OTPAuthURL))

	return png, err
}

type VerificationTokenRequestIDO struct {
	Base32     string `json:"base32"`
	OTPAuthURL string `json:"otpauth_url"`
}

type VerificationTokenResponseIDO struct {
	Base32     string `json:"base32"`
	OTPAuthURL string `json:"otpauth_url"`
}

// VerifyOTP function verifies provided token from the third-party authenticator app. The purpose of this function is to finish the otp setup.
func (impl *GatewayControllerImpl) VerifyOTP(ctx context.Context, req *VerificationTokenRequestIDO) (*VerificationTokenResponseIDO, error) {
	// Extract from our session the following data.
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	// sessionID, _ := ctx.Value(constants.SessionID).(string)

	////
	//// Start the transaction.
	////

	session, err := impl.DbClient.StartSession()
	if err != nil {
		impl.Logger.Error("start session error",
			slog.Any("error", err))
		return nil, err
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {

		// Lookup the user in our database, else return a `400 Bad Request` error.
		u, err := impl.UserStorer.GetByID(sessCtx, userID)
		if err != nil {
			impl.Logger.Error("failed getting session user", slog.Any("err", err))
			return nil, err
		}
		if u == nil {
			impl.Logger.Warn("user does not exist validation error")
			return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
		}
		if u.OTPSecret == "" {
			impl.Logger.Warn("user did not run generate otp")
			return nil, httperror.NewForBadRequestWithSingleField("message", "you did not setup two-factor authentication")
		}

		return nil, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*VerificationTokenResponseIDO), nil
}
