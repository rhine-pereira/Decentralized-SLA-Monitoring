package chain

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const contractABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"bucketId","type":"uint256"},{"indexed":false,"internalType":"bytes32","name":"root","type":"bytes32"}],"name":"RootStored","type":"event"},{"inputs":[{"internalType":"uint256","name":"bucketId","type":"uint256"}],"name":"getRoot","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"roots","outputs":[{"internalType":"bytes32","name":"root","type":"bytes32"},{"internalType":"uint256","name":"timestamp","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"bucketId","type":"uint256"},{"internalType":"bytes32","name":"root","type":"bytes32"}],"name":"storeRoot","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

type Client struct {
	ethClient       *ethclient.Client
	privateKey      *ecdsa.PrivateKey
	contractAddress common.Address
	parsedABI       abi.ABI
}

func NewClient(rpcURL, privKeyHex, contractAddr string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rpc: %v", err)
	}

	if strings.HasPrefix(privKeyHex, "0x") {
		privKeyHex = privKeyHex[2:]
	}
	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	parsed, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi: %v", err)
	}

	return &Client{
		ethClient:       client,
		privateKey:      privateKey,
		contractAddress: common.HexToAddress(contractAddr),
		parsedABI:       parsed,
	}, nil
}

func (c *Client) StoreRoot(bucketID int64, rootHashHex string) (string, error) {
	if strings.HasPrefix(rootHashHex, "0x") {
		rootHashHex = rootHashHex[2:]
	}
	rootBytes, err := hex.DecodeString(rootHashHex)
	if err != nil {
		return "", fmt.Errorf("invalid root hash: %v", err)
	}
	var root [32]byte
	copy(root[:], rootBytes)

	publicKey := c.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := c.ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	gasPrice, err := c.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	chainID, err := c.ethClient.ChainID(context.Background())
	if err != nil {
		return "", err
	}

	data, err := c.parsedABI.Pack("storeRoot", big.NewInt(bucketID), root)
	if err != nil {
		return "", err
	}

	gasLimit, err := c.ethClient.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &c.contractAddress,
		Data: data,
	})
	if err != nil {
		if strings.Contains(err.Error(), "Already exists") {
			return "ALREADY_STORED", nil
		}
		return "", fmt.Errorf("failed to estimate gas: %v", err)
	}

	tx := types.NewTransaction(nonce, c.contractAddress, big.NewInt(0), gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return "", err
	}

	err = c.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func (c *Client) GetRoot(bucketID int64) (string, error) {
	data, err := c.parsedABI.Pack("getRoot", big.NewInt(bucketID))
	if err != nil {
		return "", err
	}

	res, err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &c.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return "", err
	}

	var root [32]byte
	output, err := c.parsedABI.Unpack("getRoot", res)
	if err != nil {
		return "", err
	}
	root = *abi.ConvertType(output[0], new([32]byte)).(*[32]byte)

	return hex.EncodeToString(root[:]), nil
}
