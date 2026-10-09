package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"

	applicationapproval "github.com/balzorn/e5-atlasis/backend/internal/application/approval"
	applicationasset "github.com/balzorn/e5-atlasis/backend/internal/application/asset"
	applicationchange "github.com/balzorn/e5-atlasis/backend/internal/application/change"
	"github.com/balzorn/e5-atlasis/backend/internal/httpapi"
	"github.com/balzorn/e5-atlasis/backend/internal/infrastructure/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	db, err := postgres.New(ctx, databaseURL)
	if err != nil {
		logger.Error("initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	authMode := os.Getenv("ATLASIS_AUTH_MODE")
	principalResolver, err := principalResolverForConfig(authMode, addr)
	if err != nil {
		logger.Error("configure API authentication", "error", err)
		os.Exit(1)
	}
	if strings.TrimSpace(authMode) == "" {
		logger.Warn("ATLASIS_AUTH_MODE is unset; API routes will reject requests until trusted authentication is configured")
	}
	if strings.TrimSpace(authMode) == developmentHeaderAuthMode {
		logger.Warn("development header authentication enabled; API listener is restricted to a loopback IP address")
	}

	assetRepository := postgres.NewAssetRepository(db)
	changeRequestRepository := postgres.NewChangeRequestRepository(db)
	assetIDGenerator := postgres.NewAssetIDGenerator(db)
	changeRequestIDGenerator := postgres.NewChangeRequestIDGenerator(db)
	approvalRepository := postgres.NewApprovalRepository(db)
	approvalIDGenerator := postgres.NewApprovalIDGenerator(db)
	changeApplier := postgres.NewChangeApplier(db)

	getAsset := applicationasset.NewGetAssetUseCase(assetRepository)
	createAsset := applicationasset.NewCreateAssetUseCase(assetRepository, assetIDGenerator)
	getChangeRequest := applicationchange.NewGetChangeRequestUseCase(changeRequestRepository)
	createChangeRequest := applicationchange.NewCreateChangeRequestUseCase(
		assetRepository,
		changeRequestRepository,
		changeRequestIDGenerator,
	)
	submitChange := applicationchange.NewSubmitChangeRequestUseCase(changeRequestRepository)
	startReview := applicationchange.NewStartReviewUseCase(changeRequestRepository)
	requestChanges := applicationchange.NewRequestChangesUseCase(changeRequestRepository)
	rejectChange := applicationchange.NewRejectChangeRequestUseCase(changeRequestRepository)
	approveChange := applicationchange.NewApproveChangeRequestUseCase(changeRequestRepository, approvalRepository)
	applyChange := applicationchange.NewApplyChangeRequestUseCase(
		assetRepository,
		changeRequestRepository,
		changeApplier,
	)
	createApproval := applicationapproval.NewCreateApprovalUseCase(
		changeRequestRepository,
		approvalRepository,
		approvalIDGenerator,
	)
	listApprovals := applicationapproval.NewListApprovalsUseCase(
		changeRequestRepository,
		approvalRepository,
	)
	approveApproval := applicationapproval.NewApproveApprovalUseCase(approvalRepository)
	rejectApproval := applicationapproval.NewRejectApprovalUseCase(approvalRepository)

	handler := httpapi.NewHandler(
		getAsset,
		getChangeRequest,
		createAsset,
		createChangeRequest,
		submitChange,
		startReview,
		requestChanges,
		rejectChange,
		approveChange,
		applyChange,
		createApproval,
		listApprovals,
		approveApproval,
		rejectApproval,
		principalResolver,
	)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("HTTP server starting", "addr", addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("HTTP server stopped")
}
