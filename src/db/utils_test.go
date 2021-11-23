package db

import "testing"

type PairIS struct {
	Input  interface{}
	Output string
}

func TestObjectToSQLConditionDontPermitDefault(t *testing.T) {
	InputData := [...]PairIS{
		{User{
			Id:    4,
			Email: "232",
		}, "`Id` = ? AND `Email` = ?"},
		{User{
			Id:    0,
			Email: "232",
		}, "`Email` = ?"},
		{User{
			Email: "232",
		}, "`Email` = ?"},
	}

	for _, data := range InputData {
		output, _ := ObjectToSQLCondition(AND, data.Input, false)
		if output != data.Output {
			t.Errorf("Test failed! Input: %v, expected output: %v, real output: %v\n", data.Input, data.Output, output)
		}
	}
}

func TestObjectToSQLConditionPermitDefault(t *testing.T) {
	InputData := [...]PairIS{
		{User{
			Id:    4,
			Email: "232",
		}, "`Id` = ? AND `Email` = ? AND `CreateDate` = ?"},
		{User{
			Id:    0,
			Email: "232",
		}, "`Id` = ? AND `Email` = ? AND `CreateDate` = ?"},
		{User{
			Email: "232",
		}, "`Id` = ? AND `Email` = ? AND `CreateDate` = ?"},
	}

	for _, data := range InputData {
		output, _ := ObjectToSQLCondition(AND, data.Input, true)
		if output != data.Output {
			t.Errorf("Test failed! Input: %v, expected output: %v, real output: %v\n", data.Input, data.Output, output)
		}
	}

}
