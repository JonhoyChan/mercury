package config

const (
	ViperKeyServiceName      = "service_name"
	ViperKeyVersion          = "version"
	ViperKeyRegisterTTL      = "register_ttl"
	ViperKeyRegisterInterval = "register_interval"
	ViperKeyHost             = "host"
	ViperKeyPort             = "port"
	ViperKeyRPCPort          = "rpc_port"

	// etcd
	ViperKeyEtcdEnable    = "registry.etcd.enable"
	ViperKeyEtcdAddresses = "registry.etcd.addresses"
	ViperKeyEtcdTimeout   = "registry.etcd.timeout"

	// stan
	ViperKeyStanEnable      = "broker.stan.enable"
	ViperKeyStanAddresses   = "broker.stan.addresses"
	ViperKeyStanClusterID   = "broker.stan.cluster_id"
	ViperKeyStanDurableName = "broker.stan.durable_name"

	// log
	ViperKeyLogMode = "log.mode"

	// database
	ViperKeyDatabaseDriver      = "database.driver"
	ViperKeyDatabaseDSN         = "database.dsn"
	ViperKeyDatabaseActive      = "database.active"
	ViperKeyDatabaseIdle        = "database.idle"
	ViperKeyDatabaseIdleTimeout = "database.idle_timeout"

	// redis
	ViperKeyRedisUsername    = "redis.username"
	ViperKeyRedisAddress     = "redis.address"
	ViperKeyRedisPassword    = "redis.password"
	ViperKeyRedisDB          = "redis.db"
	ViperKeyRedisIdleTimeout = "redis.idle_timeout"

	// hasher argon2
	ViperKeyHasherArgon2Parallelism = "hasher.argon2.parallelism"
	ViperKeyHasherArgon2Memory      = "hasher.argon2.memory"
	ViperKeyHasherArgon2Iterations  = "hasher.argon2.iterations"
	ViperKeyHasherArgon2SaltLength  = "hasher.argon2.salt_length"
	ViperKeyHasherArgon2KeyLength   = "hasher.argon2.key_length"

	// generator uid
	ViperKeyGeneratorUidWorkID = "generator.uid.work_id"
	ViperKeyGeneratorUidKey    = "generator.uid.key"

	// authenticator token
	ViperKeyAuthenticatorTokenEnable       = "authenticator.token.enable"
	ViperKeyAuthenticatorTokenExpire       = "authenticator.token.expire"
	ViperKeyAuthenticatorTokenSerialNumber = "authenticator.token.serial_number"
	ViperKeyAuthenticatorTokenKey          = "authenticator.token.key"

	// authenticator jwt
	ViperKeyAuthenticatorJWTEnable       = "authenticator.jwt.enable"
	ViperKeyAuthenticatorJWTExpire       = "authenticator.jwt.expire"
	ViperKeyAuthenticatorJWTSerialNumber = "authenticator.jwt.serial_number"
	ViperKeyAuthenticatorJWTKey          = "authenticator.jwt.key"
	ViperKeyAuthenticatorJWTIssuer       = "authenticator.jwt.issuer"
)
