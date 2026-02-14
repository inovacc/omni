// Copyright 2020-2025 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bufprotosource

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type location struct {
	filePath               string
	sourceCodeInfoLocation *descriptorpb.SourceCodeInfo_Location
}

func newLocation(
	filePath string,
	sourceCodeInfoLocation *descriptorpb.SourceCodeInfo_Location,
) *location {
	return &location{
		filePath:               filePath,
		sourceCodeInfoLocation: sourceCodeInfoLocation,
	}
}

func (l *location) FilePath() string {
	return l.filePath
}

func (l *location) StartLine() int {
	switch len(l.sourceCodeInfoLocation.GetSpan()) {
	case 3, 4:
		return int(l.sourceCodeInfoLocation.GetSpan()[0]) + 1
	default:
		// since we are not erroring, making this and others 1 so that other code isn't messed up by assuming
		// this is >= 1
		return 1
	}
}

func (l *location) StartColumn() int {
	switch len(l.sourceCodeInfoLocation.GetSpan()) {
	case 3, 4:
		return int(l.sourceCodeInfoLocation.GetSpan()[1]) + 1
	default:
		// since we are not erroring, making this and others 1 so that other code isn't messed up by assuming
		// this is >= 1
		return 1
	}
}

func (l *location) EndLine() int {
	switch len(l.sourceCodeInfoLocation.GetSpan()) {
	case 3:
		return int(l.sourceCodeInfoLocation.GetSpan()[0]) + 1
	case 4:
		return int(l.sourceCodeInfoLocation.GetSpan()[2]) + 1
	default:
		// since we are not erroring, making this and others 1 so that other code isn't messed up by assuming
		// this is >= 1
		return 1
	}
}

func (l *location) EndColumn() int {
	switch len(l.sourceCodeInfoLocation.GetSpan()) {
	case 3:
		return int(l.sourceCodeInfoLocation.GetSpan()[2]) + 1
	case 4:
		return int(l.sourceCodeInfoLocation.GetSpan()[3]) + 1
	default:
		// since we are not erroring, making this and others 1 so that other code isn't messed up by assuming
		// this is >= 1
		return 1
	}
}

func (l *location) LeadingComments() string {
	if l.sourceCodeInfoLocation.LeadingComments == nil {
		return ""
	}

	return l.sourceCodeInfoLocation.GetLeadingComments()
}

func (l *location) TrailingComments() string {
	if l.sourceCodeInfoLocation.TrailingComments == nil {
		return ""
	}

	return l.sourceCodeInfoLocation.GetTrailingComments()
}

func (l *location) LeadingDetachedComments() []string {
	return l.sourceCodeInfoLocation.GetLeadingDetachedComments()
}

func (l *location) SourcePath() protoreflect.SourcePath {
	return protoreflect.SourcePath(l.sourceCodeInfoLocation.GetPath())
}
