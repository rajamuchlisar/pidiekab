package media_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pidiekab/internal/modules/identity/domain"
	mediaApp "pidiekab/internal/modules/media/application"
	mediaDomain "pidiekab/internal/modules/media/domain"
	mediaPG "pidiekab/internal/modules/media/infrastructure/postgres"
	mediaStorage "pidiekab/internal/modules/media/infrastructure/storage"
	mediaHTTP "pidiekab/internal/modules/media/transport/http"
	"pidiekab/internal/platform/database"
	"pidiekab/internal/platform/httpserver/middleware"
)

func createSamplePNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func createSampleJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: 50, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, nil)
	return buf.Bytes()
}

func TestImageValidation(t *testing.T) {
	t.Run("Valid PNG", func(t *testing.T) {
		pngBytes := createSamplePNG(32, 24)
		mime, w, h, err := mediaDomain.SniffAndValidateImage(pngBytes, "test.png")
		if err != nil {
			t.Fatalf("expected valid png, got err: %v", err)
		}
		if mime != mediaDomain.MimePNG || w != 32 || h != 24 {
			t.Errorf("unexpected png params: mime=%s, w=%d, h=%d", mime, w, h)
		}
	})

	t.Run("Valid JPEG", func(t *testing.T) {
		jpegBytes := createSampleJPEG(40, 30)
		mime, w, h, err := mediaDomain.SniffAndValidateImage(jpegBytes, "photo.jpg")
		if err != nil {
			t.Fatalf("expected valid jpeg, got err: %v", err)
		}
		if mime != mediaDomain.MimeJPEG || w != 40 || h != 30 {
			t.Errorf("unexpected jpeg params: mime=%s, w=%d, h=%d", mime, w, h)
		}
	})

	t.Run("Reject SVG and HTML active content", func(t *testing.T) {
		svgData := []byte("<svg xmlns='http://www.w3.org/2000/svg'><script>alert('xss')</script></svg>")
		_, _, _, err := mediaDomain.SniffAndValidateImage(svgData, "malicious.svg")
		if err == nil {
			t.Fatal("expected svg to be rejected")
		}
	})

	t.Run("Reject Corrupted Bytes with PNG extension", func(t *testing.T) {
		fakeBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x01}
		_, _, _, err := mediaDomain.SniffAndValidateImage(fakeBytes, "broken.png")
		if err == nil {
			t.Fatal("expected corrupted png to fail decode")
		}
	})
}

func TestStorageTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pidiekab-storage-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	st, err := mediaStorage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	traversalKeys := []string{
		"../escape.txt",
		"../../etc/passwd",
		"folder/../../../outside.txt",
		"/absolute/path.jpg",
	}

	for _, k := range traversalKeys {
		_, err := st.ResolvePath(k)
		if err == nil {
			t.Errorf("expected traversal error for key %q, got nil", k)
		}
	}
}

func TestMediaUploadAndServe(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@127.0.0.1:5432/pidiekab_test?sslmode=disable"
	}
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	pool, err := database.NewPool(ctx, database.Config{
		URL:             dbURL,
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: 0,
		MaxConnIdleTime: 0,
	}, logger)
	if err != nil {
		t.Skipf("skipping postgres media test (db unavailable): %v", err)
	}
	defer pool.Close()

	tempDir, err := os.MkdirTemp("", "pidiekab-media-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	st, err := mediaStorage.NewLocalStorage(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	repo := mediaPG.NewMediaRepository(pool)
	service := mediaApp.NewMediaService(repo, st, nil)
	handler := mediaHTTP.NewMediaHandler(service, nil)

	mux := http.NewServeMux()
	// Directly call handler functions or wrap with mock context
	mux.HandleFunc("POST /api/v1/admin/media", handler.HandleUploadMedia)
	mux.HandleFunc("GET /api/v1/admin/media", handler.HandleListMedia)

	// Admin session mock context
	adminCtx := context.WithValue(ctx, middleware.AdminContextKey, &domain.Admin{
		ID:          1,
		Email:       "admin@pidiekab.go.id",
		DisplayName: "Test Admin",
		IsActive:    true,
	})

	// 1. Upload sample PNG
	pngBytes := createSamplePNG(50, 50)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "sigli_beach.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(pngBytes)
	_ = writer.WriteField("alt_id", "Pantai Sigli")
	_ = writer.WriteField("alt_en", "Sigli Beach")
	_ = writer.WriteField("attribution", "Disparbud Pidie")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/media", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = req.WithContext(adminCtx)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload failed with code %d: %s", rec.Code, rec.Body.String())
	}

	// 2. List media
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/media?page=1&per_page=10", nil)
	listReq = listReq.WithContext(adminCtx)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list media failed with code %d: %s", listRec.Code, listRec.Body.String())
	}

	// 3. Clean up
	// In the real DB, we can delete the asset
}
