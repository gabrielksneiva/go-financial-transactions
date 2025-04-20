package client

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/crypto"      // FromECDSAPub, Keccak256, Sign
	"github.com/fbsobreira/gotron-sdk/pkg/client" // gRPC client :contentReference[oaicite:7]{index=7}
	"github.com/mr-tron/base58"                   // Base58Check :contentReference[oaicite:8]{index=8}
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/gabrielksneiva/go-financial-transactions/domain"
)

type validateRequest struct {
	Address string `json:"address"`
}
type validateResponse struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

type TronClient struct {
	grpcClient  *client.GrpcClient
	privateKey  *ecdsa.PrivateKey
	fromAddress string
}

func NewTronClient() domain.BlockchainClient {
	grpcCli := client.NewGrpcClient(os.Getenv("TRON_GRPC_URL"))
	if err := grpcCli.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		log.Fatalf("❌ Erro ao conectar gRPC TRON: %v", err)
	}
	pkHex := os.Getenv("TRON_PRIVATE_KEY")
	pk, err := crypto.HexToECDSA(pkHex) // HexToECDSA parseia chave sem 0x
	if err != nil {
		log.Fatalf("❌ TRON_PRIVATE_KEY inválida: %v", err)
	}
	return &TronClient{
		grpcClient:  grpcCli,
		privateKey:  pk,
		fromAddress: os.Getenv("TRON_FROM_ADDR"),
	}
}

func AddressFromPubKey(pub *ecdsa.PublicKey) string {
	uncompressed := crypto.FromECDSAPub(pub)   // 0x04‖X‖Y :contentReference[oaicite:9]{index=9}
	hash := crypto.Keccak256(uncompressed[1:]) // Keccak256 nas coordenadas :contentReference[oaicite:10]{index=10}
	payload := append([]byte{0x41}, hash[len(hash)-20:]...)
	h1 := sha256.Sum256(payload)
	h2 := sha256.Sum256(h1[:])
	return base58.Encode(append(payload, h2[:4]...)) // Base58Check :contentReference[oaicite:11]{index=11}
}

func (t *TronClient) SendSignedTRX(tx domain.BlockchainTransaction, transactionID string) (*domain.BlockchainTxResult, error) {
	log.Println("🚀 Iniciando envio TRX (Shasta)")

	derived := AddressFromPubKey(&t.privateKey.PublicKey)
	if derived != t.fromAddress {
		return nil, fmt.Errorf("chave privada não pertence a %s", t.fromAddress)
	}

	// 1. Cria a transação inicial via gRPC (Transfer)
	extTx, err := t.grpcClient.Transfer(t.fromAddress, tx.ToAddress, tx.Amount)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar transação: %w", err)
	}

	// 2. Injeta o ID local (UUID) no campo raw_data.data para tornar o payload único
	extTx.Transaction.RawData.Data = []byte(transactionID) // raw_data.data é campo de memo :contentReference[oaicite:3]{index=3}

	// 3. Recalcula o hash (txID) após modificar raw_data
	if err := t.grpcClient.UpdateHash(extTx); err != nil {
		return nil, fmt.Errorf("falha ao atualizar hash após injetar ID: %w", err)
	}

	// 4. Serializa o raw_data já atualizado
	rawBytes, err := proto.Marshal(extTx.Transaction.GetRawData())
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar raw_data: %w", err)
	}

	// 5. Calcula SHA-256 e gera a assinatura
	h := sha256.Sum256(rawBytes) // protocolo TRON usa SHA-256 :contentReference[oaicite:4]{index=4}
	sig, err := crypto.Sign(h[:], t.privateKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar: %w", err)
	}
	extTx.Transaction.Signature = append(extTx.Transaction.Signature, sig)

	// 6. Transmite a transação para o fullnode
	res, err := t.grpcClient.Broadcast(extTx.Transaction)
	if err != nil {
		return nil, fmt.Errorf("erro ao transmitir TX: %w", err)
	}
	if !res.Result {
		return nil, fmt.Errorf("falha no broadcast: %s", res.String())
	}

	// 7. Retorna o resultado, incluindo o novo txID
	txID := fmt.Sprintf("%x", extTx.GetTxid())
	return &domain.BlockchainTxResult{
		TxID:        txID,
		FromAddress: t.fromAddress,
		ToAddress:   tx.ToAddress,
		Amount:      float64(tx.Amount) / 1e6,
	}, nil
}

func ValidateTronAddress(address string) (bool, error) {
	url := os.Getenv("TRON_URL") + "/wallet/validateaddress"
	b, _ := json.Marshal(validateRequest{Address: address})
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return false, fmt.Errorf("erro ao validar endereço: %w", err)
	}
	defer resp.Body.Close()
	var result validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("erro ao decodificar validação: %w", err)
	}
	return result.Result, nil
}
