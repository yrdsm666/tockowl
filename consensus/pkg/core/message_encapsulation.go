package core

import (
	"log"

	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"

	"google.golang.org/protobuf/proto"
)

// Encapsulation encapsulates a message to a general type(*protobuf.Message)
func Encapsulation(messageType string, ID []byte, sender uint32, payloadMessage any) *protobuf.Message {
	var data []byte
	var err error
	switch messageType {
	case "Value":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Value))
	case "Echo":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Echo))
	case "Lock":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Lock))
	case "Finish":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Finish))
	case "Done":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Done))
	case "Halt":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Halt))
	case "PreVote":
		data, err = proto.Marshal((payloadMessage).(*protobuf.PreVote))
	case "Vote":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Vote))
	case "Best":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Best))
	case "Call":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Call))
	case "BLock":
		data, err = proto.Marshal((payloadMessage).(*protobuf.BLock))
	case "Dec":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Dec))
	case "Help":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Help))
	case "BVAL":
		data, err = proto.Marshal((payloadMessage).(*protobuf.BVAL))
	case "AUX":
		data, err = proto.Marshal((payloadMessage).(*protobuf.AUX))
	case "COIN":
		data, err = proto.Marshal((payloadMessage).(*protobuf.COIN))
	case "STOP":
		data, err = proto.Marshal((payloadMessage).(*protobuf.STOP))
	case "NGProposal":
		data, err = proto.Marshal((payloadMessage).(*protobuf.NGProposal))
	case "Received":
		data, err = proto.Marshal((payloadMessage).(*protobuf.Received))
	case "NGCallHelp":
		data, err = proto.Marshal((payloadMessage).(*protobuf.NGCallHelp))
	case "NGHelp":
		data, err = proto.Marshal((payloadMessage).(*protobuf.NGHelp))
	case "PREP":
		data, err = proto.Marshal((payloadMessage).(*protobuf.PREP))
	case "FA":
		data, err = proto.Marshal((payloadMessage).(*protobuf.FA))
	case "TockCatBest":
		data, err = proto.Marshal((payloadMessage).(*protobuf.TockCatBest))

	}

	if err != nil {
		log.Fatalln(err)
	}
	return &protobuf.Message{
		Type:   messageType,
		Id:     ID,
		Sender: sender,
		Data:   data,
	}
}

// Decapsulation decapsulates a message to it's original type
func Decapsulation(messageType string, m *protobuf.Message) any {
	switch messageType {
	case "Value":
		var payloadMessage protobuf.Value
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Echo":
		var payloadMessage protobuf.Echo
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Lock":
		var payloadMessage protobuf.Lock
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Finish":
		var payloadMessage protobuf.Finish
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Done":
		var payloadMessage protobuf.Done
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Halt":
		var payloadMessage protobuf.Halt
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "PreVote":
		var payloadMessage protobuf.PreVote
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Vote":
		var payloadMessage protobuf.Vote
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Best":
		var payloadMessage protobuf.Best
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Call":
		var payloadMessage protobuf.Call
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "BLock":
		var payloadMessage protobuf.BLock
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Help":
		var payloadMessage protobuf.Help
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Dec":
		var payloadMessage protobuf.Dec
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "BVAL":
		var payloadMessage protobuf.BVAL
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "AUX":
		var payloadMessage protobuf.AUX
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "COIN":
		var payloadMessage protobuf.COIN
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "STOP":
		var payloadMessage protobuf.STOP
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "NGProposal":
		var payloadMessage protobuf.NGProposal
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "Received":
		var payloadMessage protobuf.Received
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "NGCallHelp":
		var payloadMessage protobuf.NGCallHelp
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "NGHelp":
		var payloadMessage protobuf.NGHelp
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "PREP":
		var payloadMessage protobuf.PREP
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "FA":
		var payloadMessage protobuf.FA
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	case "TockCatBest":
		var payloadMessage protobuf.TockCatBest
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	default:
		var payloadMessage protobuf.Message
		proto.Unmarshal(m.Data, &payloadMessage)
		return &payloadMessage
	}
}
