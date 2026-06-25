package admin_test

import (
	"strings"
	"testing"

	admindto "course_server/internal/dto/admin"
	"course_server/internal/model"
)

func TestEssayExampleValidateForCreateWordCountFromBody(t *testing.T) {
	t.Parallel()
	status := model.StatusEnabled
	req := admindto.CreateEssayExampleRequest{
		Title:   "范文",
		Grade:   3,
		Body:    "<p>你好 世界</p>",
		Comment: "<p>点评</p>",
		Status:  &status,
	}
	essay, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if essay.WordCount != 4 {
		t.Fatalf("expected word count 4, got %d", essay.WordCount)
	}
	if strings.Contains(essay.Body, "script") {
		t.Fatalf("unexpected unsanitized body")
	}
}

func TestEssayExampleValidateForUpdateWordCountFromBody(t *testing.T) {
	t.Parallel()
	req := admindto.UpdateEssayExampleRequest{
		Title:   "范文",
		Grade:   3,
		Body:    "<p>abc</p>",
		Comment: "<p>点评</p>",
	}
	essay, err := req.ValidateForUpdate(model.EssayExample{Status: model.StatusEnabled})
	if err != nil {
		t.Fatalf("ValidateForUpdate: %v", err)
	}
	if essay.WordCount != 3 {
		t.Fatalf("expected word count 3, got %d", essay.WordCount)
	}
}

func TestEssayExampleValidateRejectsEmptyHTMLBody(t *testing.T) {
	t.Parallel()
	req := admindto.CreateEssayExampleRequest{
		Title:   "范文",
		Grade:   3,
		Body:    "<p><br></p>",
		Comment: "<p>点评</p>",
	}
	if _, err := req.ValidateForCreate(); err == nil {
		t.Fatal("expected error for empty html body")
	}
}

func TestEssayExampleValidateStripsScript(t *testing.T) {
	t.Parallel()
	req := admindto.CreateEssayExampleRequest{
		Title:   "范文",
		Grade:   3,
		Body:    `<p>正文</p><script>alert(1)</script>`,
		Comment: "<p>点评</p>",
	}
	essay, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if strings.Contains(essay.Body, "script") {
		t.Fatalf("script should be stripped, got %q", essay.Body)
	}
}
