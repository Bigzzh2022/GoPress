package posts

import (
	"encoding/json"
	"testing"
)

func TestUpdateInputCategoryIDPresenceAndNull(t *testing.T) {
	var omitted UpdateInput
	if err := json.Unmarshal([]byte(`{"title":"Updated"}`), &omitted); err != nil {
		t.Fatal(err)
	}
	if omitted.CategoryID.Set {
		t.Fatal("omitted category_id was marked as set")
	}

	var cleared UpdateInput
	if err := json.Unmarshal([]byte(`{"category_id":null}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if !cleared.CategoryID.Set || cleared.CategoryID.Value != nil {
		t.Fatalf("category_id:null = %+v, want explicit clear", cleared.CategoryID)
	}

	var assigned UpdateInput
	if err := json.Unmarshal([]byte(`{"category_id":7}`), &assigned); err != nil {
		t.Fatal(err)
	}
	if !assigned.CategoryID.Set || assigned.CategoryID.Value == nil || *assigned.CategoryID.Value != 7 {
		t.Fatalf("category_id:7 = %+v, want id 7", assigned.CategoryID)
	}

	var invalid UpdateInput
	if err := json.Unmarshal([]byte(`{"category_id":0}`), &invalid); err == nil {
		t.Fatal("category_id:0 was accepted")
	}
}
