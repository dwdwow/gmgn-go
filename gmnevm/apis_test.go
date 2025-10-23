package gmneth

import "testing"

func TestExactIn(t *testing.T) {
	params := ExactInParams{
		TokenInChain:    "bsc",
		TokenOutChain:   "bsc",
		TokenInAddress:  "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
		TokenOutAddress: "0x924fa68a0FC644485b8df8AbfA0A41C2e7744444",
		FromAddress:     "0x7d9471511a6c027e978adaf02014c3d2f40a0571",
		InAmount:        "1000000000000000000",
	}
	resp, err := ExactIn(params)
	if err != nil {
		t.Fatalf("failed to get exact in: %v", err)
	}
	t.Logf("exact in: %+v", resp)
}

func TestExactOut(t *testing.T) {
	params := ExactOutParams{
		TokenInChain:    "bsc",
		TokenOutChain:   "bsc",
		TokenInAddress:  "0x924fa68a0FC644485b8df8AbfA0A41C2e7744444",
		TokenOutAddress: "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
		OutAmount:       "1000000000000000000",
	}
	resp, err := ExactOut(params)
	if err != nil {
		t.Fatalf("failed to get exact out: %v", err)
	}
	t.Logf("exact out: %+v", resp)
}

func TestSlippage(t *testing.T) {
	params := SlippageParams{
		TokenAddress:   "0x924fa68a0FC644485b8df8AbfA0A41C2e7744444",
		TokenInAddress: "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
	}
	resp, err := Slippage(params)
	if err != nil {
		t.Fatalf("failed to get slippage: %v", err)
	}
	t.Logf("slippage: %+v", resp)
}

func TestGasPrice(t *testing.T) {
	resp, err := GasPrice("bsc")
	if err != nil {
		t.Fatalf("failed to get gas price: %v", err)
	}
	t.Logf("gas price: %+v", resp)
}
