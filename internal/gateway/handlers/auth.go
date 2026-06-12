package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type AuthHandler struct {
	rdb       *redis.Client
	jwtSecret string
}

func NewAuthHandler(rdb *redis.Client, jwtSecret string) *AuthHandler {
	return &AuthHandler{rdb: rdb, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Nonce(w http.ResponseWriter, r *http.Request) {
	wallet := r.URL.Query().Get("wallet")
	if wallet == "" {
		writeError(w, 400, "MISSING_WALLET", "wallet query param is required")
		return
	}

	b := make([]byte, 16)
	rand.Read(b)
	message := fmt.Sprintf("Sign in to SwapRouter: %s", hex.EncodeToString(b))

	h.rdb.Set(r.Context(), "nonce:"+wallet, message, 5*time.Minute)
	writeJSON(w, 200, map[string]string{"message": message})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Wallet    string `json:"wallet"`
		Signature string `json:"signature"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	wallet := body.Wallet
    if strings.HasPrefix(wallet, "0x") {
    wallet = strings.ToLower(wallet)
}
	message, err := h.rdb.GetDel(r.Context(), "nonce:"+wallet).Result()
	if err != nil {
		writeError(w, 401, "INVALID_NONCE", "nonce not found or expired")
		return
	}

	recovered, err := recoverAddress(message, body.Signature)
	if err != nil || recovered != wallet {
		writeError(w, 401, "INVALID_SIGNATURE", "signature verification failed")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"wallet": wallet,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, _ := token.SignedString([]byte(h.jwtSecret))
	writeJSON(w, 200, map[string]string{"token": signed})
}

func recoverAddress(message, sigHex string) (string, error) {
	sigBytes, _ := hex.DecodeString(strings.TrimPrefix(sigHex, "0x"))
	if len(sigBytes) != 65 {
		return "", fmt.Errorf("invalid signature length")
	}

	hash := crypto.Keccak256Hash([]byte(
		fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message),
	))

	sigBytes[64] -= 27 // MetaMask uses v=27/28, go-ethereum expects 0/1
	pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return "", err
	}

	return strings.ToLower(crypto.PubkeyToAddress(*pubKey).Hex()), nil
}
