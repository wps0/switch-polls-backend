package db

import (
	"testing"
	"time"
)

type PairIS struct {
	Input  interface{}
	Output string
}

func TestObjectToSQLConditionAndDontPermitDefault(t *testing.T) {
	InputData := [...]PairIS{
		{User{
			Id:         4,
			Email:      "232",
			CreateDate: time.Unix(32323, 332),
		}, "`id` = ? AND `email` = ? AND `create_date` = ?"},
		{User{
			Id:    0,
			Email: "232",
		}, "`email` = ?"},
		{User{
			Email: "232",
		}, "`email` = ?"},
	}

	for _, data := range InputData {
		output, _ := ObjectToSQLCondition(AND, data.Input, false)
		if output != data.Output {
			t.Errorf("Test failed! Input: %v, expected output: %v, real output: %v\n", data.Input, data.Output, output)
		}
	}
}

func TestObjectToSQLConditionAndPermitDefault(t *testing.T) {
	InputData := [...]PairIS{
		{User{
			Id:         4,
			Email:      "232",
			CreateDate: time.Unix(33332122, 332),
		}, "`id` = ? AND `email` = ? AND `create_date` = ?"},
		{User{
			Id:    0,
			Email: "232",
		}, "`id` = ? AND `email` = ? AND `create_date` = ?"},
		{User{
			Email: "232",
		}, "`id` = ? AND `email` = ? AND `create_date` = ?"},
		{
			Poll{
				Options: []PollOption{},
			},
			"`id` = ? AND `title` = ? AND `description` = ? AND `create_time` = ?",
		},
	}

	for _, data := range InputData {
		output, _ := ObjectToSQLCondition(AND, data.Input, true)
		if output != data.Output {
			t.Errorf("Test failed! Input: %v, expected output: %v, real output: %v\n", data.Input, data.Output, output)
		}
	}
}

func TestObjectToSQLConditionOrPermitDefault(t *testing.T) {
	InputData := [...]PairIS{
		{User{
			Id:         4,
			Email:      "232",
			CreateDate: time.Unix(10000000, 999),
		}, "`id` = ? OR `email` = ? OR `create_date` = ?"},
		{User{
			Email: "232",
		}, "`id` = ? OR `email` = ? OR `create_date` = ?"},
		{
			Poll{
				Options: []PollOption{},
			},
			"`id` = ? OR `title` = ? OR `description` = ? OR `create_time` = ?",
		},
	}

	for _, data := range InputData {
		output, _ := ObjectToSQLCondition(OR, data.Input, true)
		if output != data.Output {
			t.Errorf("Test failed! Input: %v, expected output: %v, real output: %v\n", data.Input, data.Output, output)
		}
	}
}
