// frontend/handler.go
package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gabrielksneiva/go-financial-transactions/api"
	"github.com/gabrielksneiva/go-financial-transactions/frontend/components"
	"github.com/gabrielksneiva/go-financial-transactions/frontend/views"
	"github.com/gofiber/fiber/v2"
)

// LoginPage passa a base da API ao template
func LoginPage(c *fiber.Ctx) error {
	apiBase := fmt.Sprintf("http://localhost:%s", os.Getenv("API_PORT"))
	fmt.Println(apiBase)
	var buf bytes.Buffer
	if err := views.Login(apiBase).Render(c.Context(), &buf); err != nil {
		return err
	}
	c.Type("html", "utf-8")
	return c.Send(buf.Bytes())
}

func RegisterPage(c *fiber.Ctx) error {
	apiBase := fmt.Sprintf("http://localhost:%s", os.Getenv("API_PORT"))
	var buf bytes.Buffer
	if err := views.Register(apiBase).Render(c.Context(), &buf); err != nil {
		return err
	}
	c.Type("html", "utf-8")
	return c.Send(buf.Bytes())
}

func Dashboard(c *fiber.Ctx) error {
	// 1) Extrai user_id de c.Locals
	userIDRaw := c.Locals("user_id")
	var userID uint
	switch v := userIDRaw.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	default:
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid user_id in token")
	}

	// 2) Lê token do cookie
	token := c.Cookies("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Usuário não autenticado")
	}

	// 3) Monta URL da API
	apiPort := os.Getenv("API_PORT")
	apiURL := fmt.Sprintf("http://localhost:%s/api/statement/%d", apiPort, userID)

	// 4) Cria e envia a requisição
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Erro ao criar request")
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).SendString("Erro ao chamar API")
	}
	defer resp.Body.Close()

	// 5) Verifica status da API
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).SendString("API error: " + string(body))
	}

	// 6) Decodifica e renderiza
	var txs api.StatementResponse
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Erro ao ler resposta da API")
	}

	var buf bytes.Buffer
	if err := views.Dashboard(txs).Render(c.Context(), &buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Erro ao renderizar dashboard")
	}

	c.Type("html", "utf-8")
	return c.Send(buf.Bytes())
}

func Deposit(c *fiber.Ctx) error {
	return proxyTransaction(c, "/api/deposit")
}

func Withdraw(c *fiber.Ctx) error {
	return proxyTransaction(c, "/api/withdraw")
}

// proxyTransaction lê o form, token do cookie e repassa para a API real
func proxyTransaction(c *fiber.Ctx, path string) error {
	token := c.Cookies("token")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Não autenticado")
	}

	amount := c.FormValue("amount")
	apiURL := fmt.Sprintf("http://localhost:%s%s", os.Getenv("API_PORT"), path)

	form := url.Values{}
	form.Add("amount", amount)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Erro interno")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).SendString("Falha ao chamar API")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.Status(resp.StatusCode).Send(body)
	return nil
}

// TransactionExtractPartial devolve só o HTML do extrato
func TransactionExtractPartial(c *fiber.Ctx) error {
	// pega o user_id do JWT
	userIDRaw := c.Locals("user_id")
	var userID uint
	switch v := userIDRaw.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	default:
		return c.Status(fiber.StatusUnauthorized).SendString("invalid user")
	}

	// chama a API real para buscar o statement
	apiURL := fmt.Sprintf("http://localhost:%s/api/statement/%d", os.Getenv("API_PORT"), userID)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.Cookies("token"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).SendString(err.Error())
	}
	defer resp.Body.Close()

	// desserialize e renderiza só o componente
	var txs api.StatementResponse
	json.NewDecoder(resp.Body).Decode(&txs)

	var buf bytes.Buffer
	if err := components.TransactionExtract(txs.Transactions).Render(c.Context(), &buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// devolve apenas o <div>...</div> gerado
	return c.SendString(buf.String())
}
