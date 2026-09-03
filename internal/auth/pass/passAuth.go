package pass

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/kazugmx/narubox-bot/internal/auth/pass/argon2"
	"github.com/kazugmx/narubox-bot/internal/auth/pass/bcrypt"
)

type VerifyResult struct {
	OK      bool
	NewHash *string
}

func Verify(
	chall string,
	hash string,
	argon2sv *argon2.Config,
) VerifyResult {
	fail := func() VerifyResult {
		return VerifyResult{
			false,
			nil,
		}
	}

	switch {
	case
		strings.HasPrefix(hash, "$argon2id$"):

		argon2chall, _, err := argon2sv.Verify(chall, hash)
		if err != nil {
			slog.Error("failed to verify hash with Argon2id", "error", err)
		}

		return VerifyResult{
			argon2chall, nil,
		}

	case
		strings.HasPrefix(hash, "$2a$"),
		strings.HasPrefix(hash, "$2b$"),
		strings.HasPrefix(hash, "$2y$"):

		bcryptChall := bcrypt.CheckPasswordHash(chall, hash)

		if !bcryptChall.Correct {
			slog.Warn(
				"failed to compare hash with Bcrypt",
				"error", bcryptChall.E)
		} else {
			newHash, err := argon2sv.GenerateHash(chall)
			if err != nil {
				slog.Warn("failed to generate new hash with Argon2id", "error", err)
			}
			return VerifyResult{
				true,
				&newHash,
			}
		}
	default:
		return fail()
	}

	return fail()
}

// mail rules

const mailRegexPattern = `^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

var mailRegex = regexp.MustCompile(mailRegexPattern)

func ValidateMailAddressRule(mail string) bool {
	return mailRegex.MatchString(mail)
}

// password rules

var (
	lowerRe   = regexp.MustCompile(`[a-z]`)
	upperRe   = regexp.MustCompile(`[A-Z]`)
	numberRe  = regexp.MustCompile(`[0-9]`)
	specialRe = regexp.MustCompile(`[!@#$%^&*]`)
	allowedRe = regexp.MustCompile(`^[A-Za-z0-9!@#$%^&*]+$`)
)

func ValidatePasswordRule(pass string) bool {
	n := len(pass)
	return n >= 8 &&
		n <= 128 &&
		lowerRe.MatchString(pass) &&
		upperRe.MatchString(pass) &&
		numberRe.MatchString(pass) &&
		specialRe.MatchString(pass) &&
		allowedRe.MatchString(pass)
}
