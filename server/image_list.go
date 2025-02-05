package server

import (
	"strconv"

	"github.com/cri-o/cri-o/internal/storage"
	"github.com/cri-o/cri-o/server/cri/types"
	"golang.org/x/net/context"
)

const LAYER_ANNOTATION_PREFIX = "imageLayer."

// ListImages lists existing images.
func (s *Server) ListImages(ctx context.Context, req *types.ListImagesRequest) (*types.ListImagesResponse, error) {
	filter := ""
	reqFilter := req.Filter
	if reqFilter != nil {
		filterImage := reqFilter.Image
		if filterImage != nil {
			filter = filterImage.Image
		}
	}
	results, err := s.StorageImageServer().ListImages(s.config.SystemContext, filter)
	if err != nil {
		return nil, err
	}
	resp := &types.ListImagesResponse{}
	for i := range results {
		image := ConvertImage(&results[i])
		resp.Images = append(resp.Images, image)
	}
	return resp, nil
}

// ConvertImage takes an containers/storage ImageResult and converts it into a
// CRI protobuf type. More information about the "why"s of this function can be
// found in ../cri.md.
func ConvertImage(from *storage.ImageResult) *types.Image {
	if from == nil {
		return nil
	}

	repoTags := []string{}
	repoDigests := []string{}

	if len(from.RepoTags) > 0 {
		repoTags = from.RepoTags
	}

	if len(from.RepoDigests) > 0 {
		repoDigests = from.RepoDigests
	} else if from.PreviousName != "" && from.Digest != "" {
		repoDigests = []string{from.PreviousName + "@" + string(from.Digest)}
	}

	layers := make(map[string]string)
	for key, value := range from.LayersInfo {
		if value > 0 {
			// Add prefix to differenciate our annotation from others
			layers[LAYER_ANNOTATION_PREFIX+key] = strconv.FormatUint(uint64(value), 10)
		}
	}

	to := &types.Image{
		ID:          from.ID,
		RepoTags:    repoTags,
		RepoDigests: repoDigests,
		Spec:        &types.ImageSpec{Image: from.ID, Annotations: layers},
	}

	uid, username := getUserFromImage(from.User)
	to.Username = username

	if uid != nil {
		to.UID = &types.Int64Value{Value: *uid}
	}
	if from.Size != nil {
		to.Size = *from.Size
	}

	return to
}
