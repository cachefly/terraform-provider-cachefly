## 1.2.0 (Unreleased)

BREAKING CHANGES:

* resource/cachefly_log_target: The `ELASTICSEARCH` log target type was removed from the CacheFly API and is no longer supported. The `hosts`, `ssl`, `ssl_certificate_verification`, `index`, `user`, and `api_key` attributes were removed.
* resource/cachefly_log_target: Changing `type` now forces replacement of the log target.

FEATURES:

* resource/cachefly_log_target: Added support for the new `AZURE_BLOB` (`account_name`, `account_key`, `container_name`, `endpoint_protocol`, `endpoint_suffix`, `prefix`) and `HTTP` (`uri`, `method`, `auth`, `username`, `password`, `token`) log target types.
* resource/cachefly_log_target: Added the `format`, `compression`, and `sampling` log delivery options (all log target types).
* resource/cachefly_log_target: Configurations are now validated per log target type (required fields and fields that do not apply to the chosen type are reported at plan time).
* data-source/cachefly_log_targets: Updated attributes to match the new log target types.

## 0.1.0 (Unreleased)

FEATURES:
