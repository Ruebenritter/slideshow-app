package main

// This file contains the image repository functions
// → load images from directory, filter by extension, and return a list of image paths.

import (
	"math/rand"
	"os"
	"path/filepath"
)

func isSupportedImageFile(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

func getImagesFromDir(dir string) []string {
	var images []string
	// ToDo: switch to WalkDir for better performance and less memory usage
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isSupportedImageFile(info.Name()) {
			images = append(images, path)
		}
		return nil
	})
	return images
}

func shuffleImages(images *[]string) {
	rand.Shuffle(len(*images), func(i, j int) {
		(*images)[i], (*images)[j] = (*images)[j], (*images)[i]
	})
}
