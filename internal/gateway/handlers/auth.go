package handlers

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
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
	if strings.HasPrefix(wallet, "0x") {
		wallet = strings.ToLower(wallet)
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

	var verified bool
	if strings.HasPrefix(wallet, "0x") {
		recovered, err := recoverEthAddress(message, body.Signature)
		verified = err == nil && recovered == wallet
	} else {
		verified = verifySolana(message, body.Signature, wallet)
	}

	if !verified {
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

// recoverEthAddress verifies an Ethereum ECDSA signature (MetaMask).
func recoverEthAddress(message, sigHex string) (string, error) {
	sigBytes, _ := hex.DecodeString(strings.TrimPrefix(sigHex, "0x"))
	if len(sigBytes) != 65 {
		return "", fmt.Errorf("invalid signature length")
	}

	hash := crypto.Keccak256Hash([]byte(
		fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message),
	))

	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}
	pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return "", err
	}
	return strings.ToLower(crypto.PubkeyToAddress(*pubKey).Hex()), nil
}

// verifySolana verifies a Phantom ed25519 signature.
func verifySolana(message, sigHex, walletAddress string) bool {
	pubKeyBytes := base58Decode(walletAddress)
	if len(pubKeyBytes) != 32 {
		return false
	}
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(sigHex, "0x"))
	if err != nil || len(sigBytes) != 64 {
		return false
	}
	return ed25519.Verify(pubKeyBytes, []byte(message), sigBytes)
}

const b58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Decode(input string) []byte {
	result := big.NewInt(0)
	for _, c := range input {
		idx := strings.IndexRune(b58Alphabet, c)
		if idx < 0 {
			return nil
		}
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(idx)))
	}
	decoded := result.Bytes()
	if len(decoded) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(decoded):], decoded)
		return padded
	}
	return decoded
}
