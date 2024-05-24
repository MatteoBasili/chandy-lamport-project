package utils_test

import (
	"chandy_lamport/src/utils"
	"testing"
)

func TestCreateMarkChans(t *testing.T) {
	size := 10
	markChans := utils.CreateMarkChans(size)

	if cap(markChans.SendCh) != size {
		t.Errorf("Expected SendCh channel capacity %d, but got %d", size, cap(markChans.SendCh))
	}

	if cap(markChans.RecvCh) != size {
		t.Errorf("Expected RecvCh channel capacity %d, but got %d", size, cap(markChans.RecvCh))
	}
}

func TestCreateAppMsgChans(t *testing.T) {
	size := 20
	appMsgChans := utils.CreateAppMsgChans(size)

	if cap(appMsgChans.SendToProcCh) != size {
		t.Errorf("Expected SendToProcCh channel capacity %d, but got %d", size, cap(appMsgChans.SendToProcCh))
	}

	if cap(appMsgChans.RecvCh) != size {
		t.Errorf("Expected RecvCh channel capacity %d, but got %d", size, cap(appMsgChans.RecvCh))
	}

	if cap(appMsgChans.SendToSnapCh) != size {
		t.Errorf("Expected SendToSnapCh channel capacity %d, but got %d", size, cap(appMsgChans.SendToSnapCh))
	}
}

func TestCreateStatesChans(t *testing.T) {
	size := 30
	statesChans := utils.CreateStatesChans(size)

	if cap(statesChans.SaveCh) != size {
		t.Errorf("Expected SaveCh channel capacity %d, but got %d", size, cap(statesChans.SaveCh))
	}

	if cap(statesChans.CurrCh) != size {
		t.Errorf("Expected CurrCh channel capacity %d, but got %d", size, cap(statesChans.CurrCh))
	}

	if cap(statesChans.RecvCh) != size {
		t.Errorf("Expected RecvCh channel capacity %d, but got %d", size, cap(statesChans.RecvCh))
	}
}
