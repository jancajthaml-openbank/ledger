// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"github.com/jancajthaml-openbank/ledger-unit/boot"
)

func main() {
	fmt.Println(">>> Start <<<")

	program := boot.NewProgram()
	program.Setup()

	defer func() {
		program.Stop()
		fmt.Println(">>> Stop <<<")
	}()

	ctx, cancel := context.WithCancel(context.Background())
	program.Start(ctx, cancel)
}
