package consumer

import "fmt"

func ProcessSqsMsg(message interface{}) {
	fmt.Println("[ CONSUMER ] message received", message)
}
