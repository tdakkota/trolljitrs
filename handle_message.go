package main

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

func (t *Troll) OnNewMessage(ctx context.Context, _ tg.Entities, update *tg.UpdateNewMessage) error {
	msg, ok := update.Message.(*tg.Message)
	if !ok || msg.Out {
		return nil
	}

	u, ok := msg.GetPeerID().(*tg.PeerUser)
	if !ok {
		return nil
	}

	resolved, ok := t.checkUserID(u.UserID)
	if !ok {
		return nil
	}

	t.logger.Info("Got message",
		zap.String("text", msg.Message),
		zap.Time("date", time.Unix(int64(msg.Date), 0)),
	)

	// #nosec G404
	if sticker, ok := t.checkSticker(); ok && rand.Int()%2 == 0 {
		return t.ignored(ctx, resolved, msg.ID, sticker)
	}

	return t.revoke(ctx, resolved, msg.ID)
}

func (t *Troll) ignored(ctx context.Context, resolved tg.InputPeerUser, msgID int, sticker tg.Document) error {
	t.logger.Info("Answer sticker", zap.Int("msg_id", msgID))

	_, err := t.sender.To(&resolved).
		Reply(msgID).
		Document(ctx, sticker.AsInputDocumentFileLocation())
	return err
}

func (t *Troll) revoke(ctx context.Context, resolved tg.InputPeerUser, msgID int) error {
	t.logger.Info("Delete message", zap.Int("msg_id", msgID))
	self := t.sender.Self()

	_, err := self.ForwardIDs(&resolved, msgID).Send(ctx)
	if err != nil {
		t.logger.Warn("Forward failed", zap.Error(err))
	}

	_, err = self.Revoke().Messages(ctx, msgID)
	return err
}
