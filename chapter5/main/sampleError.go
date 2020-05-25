package main

import (
	"fmt"
	"runtime/debug"
)

type MyError struct {
	Inner      error
	Message    string
	StackTrace string
	Misc       map[string]interface{}
}

func wrapError(err error, messagef string, msgArgs ...interface{}) MyError {
	return MyError{
		Inner:      err, //低水準のエラーをいつでも見れるようにする
		Message:    fmt.Sprintf(messagef, msgArgs),
		StackTrace: string(debug.Stack()),        //エラーが作られたときのスタックトレースを記録するための者
		Misc:       make(map[string]interface{}), //雑多な情報の保管場所。補助資料。
	}
}

func (err MyError) Error() string {
	return err.Message
}

//func PostReport(id string) error {
//	result , err := lowlevel.DoWork()
//	if err != nil {
//		if _, ok := err.(lowlevel.Error); ok { // エラーの形式を確認。
//			err = WrapError(err, "cannnot post reort with id %q", id) //自分のモジュール向けの負荷情報とともにやってきたエラーを包んで新しい形にする。
//		}
//		return err
//	}
//}
