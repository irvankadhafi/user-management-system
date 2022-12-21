package utils

import (
	"context"
	"google.golang.org/grpc/metadata"
)

// DumpIncomingContext converts the metadata from the incoming context to a string representation using json marshal.
func DumpIncomingContext(c context.Context) string {
	md, _ := metadata.FromIncomingContext(c)
	return Dump(md)
}
