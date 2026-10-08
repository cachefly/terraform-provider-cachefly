## 1.3.0 (October 8, 2026)

FEATURES:

* **New Resource:** `cachefly_edge_control_script` publishes the edge control script of a service (`REQUEST` or `RESPONSE`) as a new version and activates it, or deactivates it with `activated = false`.
* **New Resource:** `cachefly_edge_control_kv` manages the account-level or a service-level edge control key/value store, with string, number and boolean values.
* **New Resource:** `cachefly_edge_control_library_script` manages scripts in the account's edge control script library.
* **New Data Source:** `cachefly_edge_control_script` reads the active or a specific version of an edge control script.
* **New Data Source:** `cachefly_edge_control_script_versions` lists the published versions of an edge control script.
* **New Data Source:** `cachefly_edge_control_kv` reads the account-level, a service-level or the merged edge control key/value store.
* **New Data Source:** `cachefly_edge_control_library_script` reads a `SYSTEM` or `USER` script from the edge control script library.
* **New Data Source:** `cachefly_edge_control_library_scripts` lists the edge control script library, filtered by type, kind or name.

BUG FIXES:

* resource/cachefly_service: Setting `auto_ssl = false` now disables Auto SSL. Previously `false` was dropped from the API request, so the apply failed with "Provider produced inconsistent result after apply". Services whose configuration sets `auto_ssl = false` while Auto SSL is enabled will have it disabled on the next apply.
* resource/cachefly_service: When `auto_ssl` is not set in the configuration, plans now keep its current value instead of showing it as known after apply.
* resource/cachefly_script_config: Fixed perpetual drift on `value` and "Provider produced inconsistent result after apply" errors. Values the API returns as strings are no longer JSON-encoded a second time, and values whose content matches the configuration are kept as written even when the API returns them with different whitespace or key order.

## 1.2.0 (August 6, 2026)

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
