# Go epub Tools 
this package is a simple library to change metadata of epub and get simple data like title ,author, description,date but also to edit them.

## Installation 
To install run the following command 

```bash
go install github.com/SuperninjaXII/goEpubTools/internal
```
## Example 
this simple example shows the usage in every way 

```go
package main

import (
	"fmt"
	"log"

	"github.com/SuperninjaXII/goEpubTools/internal"
)

func main() {
	// Input and output EPUB file paths
	inputEpub := "./children.epub"
	outputEpub := "updated_book.epub"

	// Step 1: Read existing metadata
	metadataPkg, err := internal.ReadMetaData(inputEpub)
	if err != nil {
		log.Fatalf("Failed to read metadata: %v", err)
	}
	if metadataPkg != nil {
		fmt.Println("=== Existing Metadata ===")
		fmt.Println("Title:", metadataPkg.MetaData.Title)
		fmt.Println("Author:", metadataPkg.MetaData.Author)
		fmt.Println("Date:", metadataPkg.MetaData.Date)
		fmt.Println("Description:", metadataPkg.MetaData.Description)
	} else {
		fmt.Println("No metadata found.")
	}

	// Step 2: Modify the metadata
	newMetadata := &internal.Package{
		MetaData: internal.MetaData{
			Title:       "New Title Here",
			Author:      "SuperninjaX2",
			Date:        "2025-07-17",
			Description: "This is an updated description of the EPUB.",
		},
	}

	// Step 3: Edit and save the updated EPUB
	fmt.Println("\n📦 Updating metadata and repacking EPUB...")
	internal.EditEpub(newMetadata, inputEpub, outputEpub)

	fmt.Println("✅ Metadata updated and saved to:", outputEpub)
}
```