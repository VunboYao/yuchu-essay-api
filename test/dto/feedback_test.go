package admindto_test

import (
	"strings"
	"testing"

	admin "course_server/internal/dto/admin"
)

func TestReplyFeedbackRequestNormalizeReply(t *testing.T) {
	req := admin.ReplyFeedbackRequest{Reply: "  感谢反馈  "}
	reply, err := req.NormalizeReply()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "感谢反馈" {
		t.Fatalf("got %q", reply)
	}
}

func TestReplyFeedbackRequestNormalizeReplyEmpty(t *testing.T) {
	req := admin.ReplyFeedbackRequest{Reply: "   "}
	if _, err := req.NormalizeReply(); err == nil {
		t.Fatal("expected error for empty reply")
	}
}

func TestReplyFeedbackRequestNormalizeReplyTooLong(t *testing.T) {
	req := admin.ReplyFeedbackRequest{Reply: strings.Repeat("a", 1001)}
	if _, err := req.NormalizeReply(); err == nil {
		t.Fatal("expected error for long reply")
	}
}
