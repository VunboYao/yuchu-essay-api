package admin_test

import (
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
		Body:    "你好 世界\n",
		Comment: "点评",
		Status:  &status,
	}
	essay, err := req.ValidateForCreate()
	if err != nil {
		t.Fatalf("ValidateForCreate: %v", err)
	}
	if essay.WordCount != 4 {
		t.Fatalf("expected word count 4, got %d", essay.WordCount)
	}
}

func TestEssayExampleValidateForUpdateWordCountFromBody(t *testing.T) {
	t.Parallel()
	req := admindto.UpdateEssayExampleRequest{
		Title:   "范文",
		Grade:   3,
		Body:    "abc",
		Comment: "点评",
	}
	essay, err := req.ValidateForUpdate(model.EssayExample{Status: model.StatusEnabled})
	if err != nil {
		t.Fatalf("ValidateForUpdate: %v", err)
	}
	if essay.WordCount != 3 {
		t.Fatalf("expected word count 3, got %d", essay.WordCount)
	}
}
