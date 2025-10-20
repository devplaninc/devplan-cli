package recipes

import (
	"context"

	"github.com/devplaninc/adcp-core/adcp/core"
	"github.com/devplaninc/adcp-core/adcp/core/executable"
	"github.com/devplaninc/adcp/clients/go/adcp"
	"github.com/devplaninc/devplan-cli/internal/devplan"
)

type PersistOptions struct {
	Path       string
	EntryPoint *adcp.EntryPoint
}

func PersistForTask(ctx context.Context, companyID int32, taskID string, opts PersistOptions) error {
	cl := devplan.NewClient(devplan.Config{})
	recipe, err := cl.GetTaskRecipe(companyID, taskID)
	if err != nil {
		return err
	}
	executableRecipe := adcp.ExecutableRecipe_builder{
		Recipe:     recipe,
		EntryPoint: opts.EntryPoint,
	}.Build()
	return persist(ctx, executableRecipe, opts.Path)
}

func persist(ctx context.Context, executableRecipe *adcp.ExecutableRecipe, path string) error {
	r := executable.ForRecipe(executableRecipe)
	res, err := r.Materialize(ctx)
	if err != nil {
		return err
	}
	return core.PersistMaterializedResult(ctx, path, res)
}
