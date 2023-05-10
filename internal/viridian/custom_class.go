package viridian

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
)

func (a API) ListCustomClasses(ctx context.Context, cluster string) ([]CustomClass, error) {
	cID, err := a.findClusterID(ctx, cluster)
	if err != nil {
		return nil, err
	}
	csw, err := doGet[[]CustomClass](ctx, fmt.Sprintf("/cluster/%s/custom_classes", cID), a.Token())
	if err != nil {
		return nil, fmt.Errorf("listing custom classes: %w", err)
	}
	return csw, nil
}

func (a API) UploadCustomClasses(ctx context.Context, sp clc.Spinner, cluster, filePath string) error {
	cID, err := a.findClusterID(ctx, cluster)
	if err != nil {
		return err
	}
	err = doCustomClassUpload(ctx, sp, fmt.Sprintf("/cluster/%s/custom_classes", cID), filePath, a.Token())
	if err != nil {
		return fmt.Errorf("uploading custom class: %w", err)
	}
	return nil
}

func (a API) DownloadCustomClass(ctx context.Context, sp clc.Spinner, targetInfo TargetInfo, cluster, artifact string) error {
	cID, err := a.findClusterID(ctx, cluster)
	if err != nil {
		return err
	}
	artifactID, artifactName, err := a.findArtifactIDAndName(ctx, cluster, artifact)
	if err != nil {
		return err
	}
	if artifactID == 0 {
		return fmt.Errorf("no such custom class found with name %d in cluster %s", artifactID, cID)
	}
	err = doCustomClassDownload(ctx, sp, targetInfo, fmt.Sprintf("/cluster/%s/custom_classes/%d", cID, artifactID), artifactName, a.token)
	if err != nil {
		return err
	}
	return nil
}

func (a API) DeleteCustomClass(ctx context.Context, cluster string, artifactID int64) error {
	cID, err := a.findClusterID(ctx, cluster)
	if err != nil {
		return err
	}
	customClasses, err := a.ListCustomClasses(ctx, cID)
	if err != nil {
		return err
	}
	var id int64
	for _, c := range customClasses {
		if c.Id == artifactID {
			id = c.Id
		}
	}
	if id == 0 {
		return fmt.Errorf("no such custom class found with name %d in cluster %s", artifactID, cluster)
	}
	err = doDelete(ctx, fmt.Sprintf("/cluster/%s/custom_classes/%d", cluster, id), a.token)
	if err != nil {
		return err
	}
	return nil
}

func (a API) findClusterID(ctx context.Context, cluster string) (string, error) {
	clusters, err := a.ListClusters(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range clusters {
		if c.ID == cluster || c.Name == cluster {
			return c.ID, nil
		}
	}
	return "", fmt.Errorf("no such class found: %s", cluster)
}

func (a API) findArtifactIDAndName(ctx context.Context, clusterName, artifact string) (int64, string, error) {
	customClasses, err := a.ListCustomClasses(ctx, clusterName)
	if err != nil {
		return 0, "", err
	}
	var artifactName string
	var artifactID int64
	for _, cc := range customClasses {
		if cc.Name == artifact || strconv.FormatInt(cc.Id, 10) == artifact {
			artifactName = cc.Name
			artifactID = cc.Id
			break
		}
	}
	return artifactID, artifactName, nil
}
