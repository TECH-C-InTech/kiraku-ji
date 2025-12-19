package main

import (
	"context"
	"fmt"
	"log"

	"backend/internal/app"
	drawdomain "backend/internal/domain/draw"
	"backend/internal/domain/post"

	"cloud.google.com/go/firestore"
)

// seedPost は Firestore posts コレクションへ投入するサンプルデータ。
type seedPost struct {
	ID      string
	Content string
	Ready   bool
}

// seedDraw は Firestore draws コレクションへ投入するサンプルデータ。
type seedDraw struct {
	PostID   string
	Result   string
	Verified bool
}

func main() {
	ctx := context.Background()

	infra, err := app.NewInfra(ctx)
	if err != nil {
		log.Fatalf("init infra: %v", err)
	}
	defer func() {
		if err := infra.Close(); err != nil {
			log.Printf("close infra: %v", err)
		}
	}()

	client := infra.Firestore()
	if client == nil {
		log.Fatalf("firestore client is not initialized; set GOOGLE_CLOUD_PROJECT and related env vars")
	}

	if err := seedFirestore(ctx, client); err != nil {
		log.Fatalf("seed firestore: %v", err)
	}

	log.Println("firestore seeding completed successfully")
}

// seedFirestore は posts / draws へ初期データを投入する。
func seedFirestore(ctx context.Context, client *firestore.Client) error {
	posts := []seedPost{
		{ID: "post-alpha", Content: "先が見えずに彷徨っている", Ready: false},
		{ID: "post-beta", Content: "光が差す日を待っている", Ready: true},
		{ID: "post-gamma", Content: "心の叫びを誰かに聞いてほしい", Ready: true},
	}

	draws := []seedDraw{
		{PostID: "post-beta", Result: "夜明けはすぐそばにあります", Verified: true},
		{PostID: "post-gamma", Result: "焦らずとも道は開けるでしょう", Verified: true},
	}

	if err := seedPosts(ctx, client, posts); err != nil {
		return err
	}
	if err := seedDraws(ctx, client, draws); err != nil {
		return err
	}
	return nil
}

// seedPosts は posts コレクションにドキュメントを保存する。
func seedPosts(ctx context.Context, client *firestore.Client, posts []seedPost) error {
	// 1件ずつ取り出して処理する
	for _, p := range posts {
		postEntity, err := post.New(post.DarkPostID(p.ID), post.DarkContent(p.Content))
		if err != nil {
			return fmt.Errorf("build post %s: %w", p.ID, err)
		}
		if p.Ready {
			if err := postEntity.MarkReady(); err != nil {
				return fmt.Errorf("mark ready %s: %w", p.ID, err)
			}
		}

		// Firestore に upsert するフィールド群。
		data := map[string]any{
			"post_id":    string(postEntity.ID()),
			"content":    string(postEntity.Content()),
			"status":     string(postEntity.Status()),
			"created_at": firestore.ServerTimestamp,
			"updated_at": firestore.ServerTimestamp,
		}

		if _, err := client.Collection("posts").Doc(p.ID).Set(ctx, data); err != nil {
			return fmt.Errorf("set post document %s: %w", p.ID, err)
		}
	}
	return nil
}

// seedDraws は draws コレクションに検証済みの結果を保存する。
func seedDraws(ctx context.Context, client *firestore.Client, draws []seedDraw) error {
	for _, d := range draws {
		drawEntity, err := drawdomain.New(post.DarkPostID(d.PostID), drawdomain.FormattedContent(d.Result))
		if err != nil {
			return fmt.Errorf("build draw %s: %w", d.PostID, err)
		}
		if d.Verified {
			drawEntity.MarkVerified()
		}

		data := map[string]any{
			"post_id":    string(drawEntity.PostID()),
			"result":     string(drawEntity.Result()),
			"status":     string(drawEntity.Status()),
			"created_at": firestore.ServerTimestamp,
		}

		if _, err := client.Collection("draws").Doc(d.PostID).Set(ctx, data); err != nil {
			return fmt.Errorf("set draw document %s: %w", d.PostID, err)
		}
	}
	return nil
}
