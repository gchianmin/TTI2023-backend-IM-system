package main

import (
	"context"
	"strings"
	"fmt"
	"time"
	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	timestamp := time.Now().Unix()

    message := &Message{
       Message:   req.Message.GetText(),
       Sender:    req.Message.GetSender(),
       Timestamp: timestamp,
    }

	roomID, err := getRoomID(req.Message.GetChat())
	if err != nil {
		return nil, err
	 }

    err = redisClient.SaveMessage(ctx, roomID, message)
    if err != nil {
       return nil, err
    }

	resp := rpc.NewSendResponse()
	resp.Code, resp.Msg = 0, "success"
	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	roomID, err := getRoomID(req.GetChat())
	if err != nil {
		return nil, err
	 }
	
	limit := int64(req.GetLimit())

	if limit == 0 {
		limit = 10 
	}
	start := req.GetCursor()
	end := start + limit // did not minus 1 on purpose for hasMore check later on

	messages, err := redisClient.GetMessagesByRoomID(ctx, roomID, start, end, req.GetReverse())
    if err != nil {
       return nil, err
    }
	respMessages := make([]*rpc.Message, 0)
    var counter int32 = 0
    var nextCursor int64 = 0
    hasMore := false
    for _, msg := range messages {
       if counter+1 > req.GetLimit() {
          // having extra value here means it has more data
          hasMore = true
          nextCursor = end
          break // do not return the last message
       }
       temp := &rpc.Message{
          Chat:     req.GetChat(),
          Text:     msg.Message,
          Sender:   msg.Sender,
          SendTime: msg.Timestamp,
       }
       respMessages = append(respMessages, temp)
       counter += 1
    }
	resp := rpc.NewPullResponse()
	resp.Messages = respMessages
    resp.Code = 0
    resp.Msg = "success"
    resp.HasMore = &hasMore
    resp.NextCursor = &nextCursor
	return resp, nil
}

// sort the sender in asc order to standardize a1:a2 and a2:a1
func getRoomID(chat string) (string, error) {
	var roomID string
	lowercase := strings.ToLower(chat)
	senders := strings.Split(lowercase, ":")

    if len(senders) != 2 {
        err := fmt.Errorf("invalid Chat ID '%s', should be in the format of user1:user2", chat)
		return "", err
    }

    sender1, sender2 := senders[0], senders[1]
    roomID = fmt.Sprintf("%s:%s", sender1, sender2)
    if strings.Compare(sender1, sender2) > 0 {
        roomID = fmt.Sprintf("%s:%s", sender2, sender1)
    }

    return roomID, nil
}
