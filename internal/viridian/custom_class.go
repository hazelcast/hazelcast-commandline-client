package viridian

import (
	"context"
	"fmt"
)

func (a API) ListCustomClasses(ctx context.Context, clusterName string) ([]CustomClass, error) {
	csw, err := doGet[[]CustomClass](ctx, fmt.Sprintf("/cluster/%s/custom_classes", clusterName), a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing custom classes: %w", err)
	}
	return csw, nil
}

func (a API) UploadCustomClasses(ctx context.Context, clusterName, filePath string) error {
	err := doCustomClassUpload(ctx, fmt.Sprintf("/cluster/%s/custom_classes", clusterName), filePath, a.Token())
	if err != nil {
		return fmt.Errorf("uploading custom class: %w", err)
	}
	return nil
}
