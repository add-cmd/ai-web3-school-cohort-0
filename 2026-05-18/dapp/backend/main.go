package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ─── 配置 ─────────────────────────────────────────────────────

var (
	rpcURL      = getEnv("RPC_URL", "https://ethereum-sepolia-rpc.publicnode.com")
	serverPort  = getEnv("PORT", "8080")
	contractABI = `[{"inputs":[],"name":"count","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"increment","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"decrement","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"newCount","type":"uint256"}],"name":"setCount","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"newCount","type":"uint256"},{"indexed":false,"internalType":"address","name":"triggeredBy","type":"address"}],"name":"CountChanged","type":"event"}]`
)

// ─── 模型 ─────────────────────────────────────────────────────

type ContractInfo struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Count   uint64 `json:"count"`
	Owner   string `json:"owner"`
}

type ApiError struct {
	Error string `json:"error"`
}

// ─── 工具 ─────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ─── 链上交互 ─────────────────────────────────────────────────

func callContract(method string, args []interface{}) ([]byte, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接 RPC 失败: %w", err)
	}
	defer client.Close()

	contractAddr := common.HexToAddress(os.Getenv("CONTRACT_ADDRESS"))

	// 使用 viem 风格的 ABI 解析
	data, err := abiEncode(method, args)
	if err != nil {
		return nil, fmt.Errorf("ABI 编码失败: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("合约调用失败: %w", err)
	}
	return result, nil
}

// 简易 ABI 编码（仅支持 view 函数，无参数或 uint256 参数）
func abiEncode(method string, args []interface{}) ([]byte, error) {
	// 函数选择器: keccak256("functionName(type1,type2,...)") 前4字节
	var sigHash string
	switch method {
	case "count":
		sigHash = "06661abd" // keccak256("count()") 前4字节
	case "owner":
		sigHash = "8da5cb5b" // keccak256("owner()") 前4字节
	default:
		return nil, fmt.Errorf("不支持的方法: %s", method)
	}

	data := common.Hex2Bytes(sigHash)

	// 如果参数是 uint256，追加 32 字节编码
	for _, arg := range args {
		switch v := arg.(type) {
		case *big.Int:
			// 填充到 32 字节
			padded := common.LeftPadBytes(v.Bytes(), 32)
			data = append(data, padded...)
		case uint64:
			padded := common.LeftPadBytes(new(big.Int).SetUint64(v).Bytes(), 32)
			data = append(data, padded...)
		default:
			return nil, fmt.Errorf("不支持的类型: %T", arg)
		}
	}
	return data, nil
}

// 简易 ABI 解码（仅支持 uint256 返回值）
func abiDecodeUint256(data []byte) (*big.Int, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("数据太短: %d bytes", len(data))
	}
	return new(big.Int).SetBytes(data[:32]), nil
}

// ─── API 路由 ─────────────────────────────────────────────────

func handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]interface{}{
		"status":  "ok",
		"network": "sepolia",
		"rpc":     rpcURL,
	})
}

func handleContractInfo(w http.ResponseWriter, r *http.Request) {
	addr := os.Getenv("CONTRACT_ADDRESS")
	if addr == "" {
		writeJSON(w, 400, ApiError{Error: "未设置 CONTRACT_ADDRESS 环境变量"})
		return
	}

	// 读取 count
	countBytes, err := callContract("count", nil)
	if err != nil {
		writeJSON(w, 500, ApiError{Error: fmt.Sprintf("读取 count 失败: %v", err)})
		return
	}
	countBig, _ := abiDecodeUint256(countBytes)

	// 读取 owner
	ownerBytes, err := callContract("owner", nil)
	if err != nil {
		writeJSON(w, 500, ApiError{Error: fmt.Sprintf("读取 owner 失败: %v", err)})
		return
	}
	owner := common.BytesToAddress(ownerBytes).Hex()

	info := ContractInfo{
		Address: addr,
		Network: "sepolia",
		Count:   countBig.Uint64(),
		Owner:   owner,
	}
	writeJSON(w, 200, info)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// ─── 主函数 ───────────────────────────────────────────────────

func main() {
	addr := os.Getenv("CONTRACT_ADDRESS")
	if addr != "" {
		log.Printf("📄 合约地址: %s", addr)
		log.Printf("🌐 RPC: %s", rpcURL)
	} else {
		log.Println("⚠ 未设置 CONTRACT_ADDRESS，/api/contract 将返回错误")
		log.Println("   部署合约后设置: export CONTRACT_ADDRESS=0x...")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/contract", handleContractInfo)
	mux.HandleFunc("/api/health", handleHealth)

	// CORS 中间件
	handler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	log.Printf("🚀 后端启动在 :%s", serverPort)
	log.Printf("   API: http://localhost:%s/api/status", serverPort)
	log.Printf("   API: http://localhost:%s/api/contract", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, handler(mux)))
}
