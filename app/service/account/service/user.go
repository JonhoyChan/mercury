package service

import (
	"context"
)

// 新用户注册
func (s *Service) Register(ctx context.Context, mobile, ip string) (string, string, error) {
	s.log.Info("[Register] request is received")

	// TODO 默认头像
	avatar := ""
	uid := s.uidGen.Get()
	// Cockroach不支持uint64类型
	id := s.uidGen.DecodeUid(uid)
	oid, err := s.persister.User().Register(ctx, id, uid, mobile, avatar, ip)
	if err != nil {
		return "", "", err
	}

	return uid.UID(), oid, nil
}

// 用户通过手机号码登录
func (s *Service) LoginViaMobile(ctx context.Context, mobile, captcha, password, version, deviceID, ip string) (string, string, error) {
	uc, err := s.persister.User().GetCredentialViaMobile(mobile)
	if err != nil {
		return "", "", err
	}

	if captcha != "" {

	} else {
		if err := s.persister.User().LoginViaPassword(ctx, uc, password, version, deviceID, ip); err != nil {
			return "", "", err
		}
	}

	uid := s.uidGen.EncodeInt64(uc.ID)
	return uid.UID(), uc.OID, nil
}

// 用户通过VID登录
func (s *Service) LoginViaVID(ctx context.Context, vid, password string) {

}
