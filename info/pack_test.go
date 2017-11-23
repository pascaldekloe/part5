package info

import (
	"encoding/hex"
	"testing"
)

func TestAddSingle(t *testing.T) {
	u := NewASDU(Narrow, GlobalAddr, M_SP_NA_1, Spont|TestFlag)
	if err := u.AddSingle(17, On|Blocked|Substituted); err != nil {
		t.Fatal("single[0] error:", err)
	}
	if err := u.AddSingle(19, Off|NotTopical|Invalid); err != nil {
		t.Fatal("single[1] error:", err)
	}

	bytes, err := u.MarshalBinary()
	if err != nil {
		t.Fatal("marshal error:", err)
	}
	if got, want := hex.EncodeToString(bytes), "010283ff113113c0"; got != want {
		t.Errorf("got 0x%s, want 0x%s", got, want)
	}
}

func TestSetSingles(t *testing.T) {
	u := NewASDU(Wide, GlobalAddr, M_SP_NA_1, Spont|TestFlag)
	u.Orig = 42
	if err := u.SetSingles(17, On, Off); err != nil {
		t.Error("set error:", err)
	}

	bytes, err := u.MarshalBinary()
	if err != nil {
		t.Fatal("marshal error:", err)
	}
	if got, want := hex.EncodeToString(bytes), "0182832affff1100000100"; got != want {
		t.Errorf("got 0x%s, want 0x%s", got, want)
	}
}

func TestAddFloat(t *testing.T) {
	u := NewASDU(Narrow, 3, M_ME_NC_1, Spont|TestFlag)
	if err := u.AddFloat(16, MeasuredFloat{99, Overflow | Invalid}); err != nil {
		t.Fatal("float[0] error:", err)
	}
	if err := u.AddFloat(17, MeasuredFloat{-1000000000, OK}); err != nil {
		t.Fatal("float[1] error:", err)
	}

	got, err := u.MarshalBinary()
	if err != nil {
		t.Fatal("marshal error:", err)
	}
	want := "0d028303100000c6428111286b6ece00"
	if s := hex.EncodeToString(got); s != want {
		t.Errorf("got 0x%s, want 0x%s", s, want)
	}
}

func TestAddFloatSetpoint(t *testing.T) {
	u := NewASDU(Narrow, 3, C_SE_NC_1, Spont|TestFlag)
	if err := u.AddFloatSetpoint(16, NewSetpointCmd(7, true), 99); err != nil {
		t.Fatal("setpoint[0] error:", err)
	}
	if err := u.AddFloatSetpoint(17, NewSetpointCmd(4, false), -1000000000); err != nil {
		t.Fatal("setpoint[1] error:", err)
	}

	got, err := u.MarshalBinary()
	if err != nil {
		t.Fatal("marshal error:", err)
	}
	want := "32028303100000c6420711286b6ece84"
	if s := hex.EncodeToString(got); s != want {
		t.Errorf("got 0x%s, want 0x%s", s, want)
	}
}
