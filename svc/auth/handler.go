package Auth

import (
	"encoding/json"
	"log"

	jwtOperator "github.com/Kazugmx/narubox-bot/internal/auth"
	"github.com/gofiber/fiber/v3"
)

func loginHandler(c fiber.Ctx, jwtEngine *jwtOperator.JWTService) error {
	req := c.Req().Body()
	var login_request UserLoginPayload
	err := json.Unmarshal(req, &login_request)
	if err != nil {
		log.Println("error:", err)
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"error": "Invalid Request payload.",
		})
	}

	token, err := jwtEngine.GenerateJwtToken(1)
	if err != nil {
		log.Println("error generating token:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token.",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func tokenCheckHandler(c fiber.Ctx, jwtEngine *jwtOperator.JWTService) error {
	jwtEngine.GenerateJwtToken(20)

	return c.JSON(fiber.Map{
		"status": "successfuly executed tokenCheckHandler",
	})
}
