package ipfs

import (
	"context"
	"fmt"
	ifacePath "github.com/ipfs/interface-go-ipfs-core/path"
	"strings"

	"github.com/ipfs/interface-go-ipfs-core/options"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
)

// UnPinDir removes all content from the published root directory to be
// garbage collected later by IPFS.
func UnPinDir(n *core.IpfsNode, rootHash string) error {
	// attempt to properly handle variations on rootHash
	if !strings.HasPrefix(rootHash, "/ipfs/") {
		rootHash = fmt.Sprintf("/ipfs/%s", rootHash)
	}
	if !strings.HasPrefix(rootHash, "/ipfs") && strings.HasPrefix(rootHash, "/") {
		rootHash = fmt.Sprintf("/ipfs%s", rootHash)
	}

	api, err := coreapi.NewCoreAPI(n)
	if err != nil {
		return fmt.Errorf("ipfs api: %s", err)
	}
	pth := ifacePath.New(rootHash)

	rp, err := api.ResolvePath(context.Background(), pth)
	if err != nil {
		return fmt.Errorf("resolve path (%s): %s", pth, err)
	}

	return api.Pin().Rm(context.Background(), rp, options.Pin.RmRecursive(true))
}
