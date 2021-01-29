package service

import (
	"bytes"
	"github.com/disintegration/imaging"
	"image"
)

//func getImageAttributes(targetWidth, targetHeight, imgWidth, imgHeight int) (width, height int) {
//	targetRatio := float32(targetWidth) / float32(targetHeight)
//	imageRatio := float32(imgWidth) / float32(imgHeight)
//	var h, w float32
//	if imageRatio > targetRatio {
//		h = float32(targetHeight)
//		w = float32(targetHeight) * imageRatio
//	} else {
//		w = float32(targetWidth)
//		h = float32(targetWidth) * (float32(imgHeight) / float32(imgWidth))
//	}
//	return int(w), int(h)
//}

func decodeImageData(imageData []byte) (image.Image, error) {
	reader := bytes.NewReader(imageData)
	img, err := imaging.Decode(reader, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	return img, err
}

//func (s *Service) resizeImage(img image.Image, imgPath, filename string, baseWidth, baseHeight int) (string, string, string, string, error) {
//	tiny, err := s.addResizedImage(img, 1*baseWidth, 1*baseHeight, path.Join(imgPath, "tiny", filename))
//	if err != nil {
//		return "", "", "", "", err
//	}
//	small, err := s.addResizedImage(img, 2*baseWidth, 2*baseHeight, path.Join(imgPath, "small", filename))
//	if err != nil {
//		return "", "", "", "", err
//	}
//	medium, err := s.addResizedImage(img, 4*baseWidth, 4*baseHeight, path.Join(imgPath, "medium", filename))
//	if err != nil {
//		return "", "", "", "", err
//	}
//	large, err := s.addResizedImage(img, 8*baseWidth, 8*baseHeight, path.Join(imgPath, "large", filename))
//	if err != nil {
//		return "", "", "", "", err
//	}
//
//	return tiny, small, medium, large, nil
//}

//func (s *Service) addImage(img image.Image, imgPath string) (string, error) {
//	out, err := os.Create(imgPath)
//	if err != nil {
//		return "", err
//	}
//	err = jpeg.Encode(out, img, nil)
//	if err != nil {
//		return "", err
//	}
//	out.Close()
//	return ipfs.AddFile(s.ipfsNode, imgPath)
//}

//func (s *Service) addResizedImage(img image.Image, w, h int, imgPath string) (string, error) {
//	width, height := getImageAttributes(w, h, img.Bounds().Max.X, img.Bounds().Max.Y)
//	newImg := imaging.Resize(img, width, height, imaging.Lanczos)
//	return s.addImage(newImg, imgPath)
//}
