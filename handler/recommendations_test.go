package handler_test

import (
	"testing"

	"github.com/oullin/handler"
)

func TestRecommendationsHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewRecommendationsHandler(f) },
		Endpoint: "/recommendations",
		Fixture:  "../storage/fixture/recommendations.json",
		Assert:   handler.AssertFirstUUID("1f58646c-ba87-4306-905f-cb64a5e49b5e"),
	})
}
