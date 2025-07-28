# Bzlmod Migration Guide

This project has been migrated from the legacy WORKSPACE system to the modern **Bzlmod** (Bazel modules) system.

## What Changed

### Before (WORKSPACE)
```starlark
workspace(name = "qnap_display_control")

http_archive(
    name = "rules_rust",
    # ... complex configuration
)

crates_repository(
    name = "crate_index",
    # ... dependency management
)
```

### After (MODULE.bazel)
```starlark
module(name = "qnap_display_control", version = "0.1.0")

bazel_dep(name = "rules_rust", version = "0.49.3")

crate = use_extension("@rules_rust//crate_universe:extension.bzl", "crate")
# ... simplified dependency management
```

## Benefits of Bzlmod

1. **Simplified Configuration**: Less boilerplate code
2. **Version Management**: Explicit versioning and compatibility
3. **Transitive Dependencies**: Better dependency resolution
4. **Future-Proof**: Official replacement for WORKSPACE
5. **Module Registry**: Centralized module discovery

## Build Configurations

The project now includes `.bazelrc` with predefined configurations:

```bash
# Optimized production build
bazel build --config=opt //...

# Static binary for deployment
bazel build --config=static //...

# Debug build with symbols
bazel build --config=debug //...

# Fast development builds
bazel build --config=dev //...
```

## Migration Steps Completed

1. ✅ Created `MODULE.bazel` to replace `WORKSPACE`
2. ✅ Updated dependency references from `@crate_index//` to `@crates//`
3. ✅ Added `.bazelrc` with build configurations
4. ✅ Created `.bazelignore` for excluded files
5. ✅ Updated installation scripts and documentation
6. ✅ Removed deprecated `WORKSPACE` file

## Compatibility

- **Minimum Bazel Version**: 6.0+ (for Bzlmod support)
- **Recommended**: Bazel 7.0+ or Bazelisk (latest)
- **Legacy Support**: None (WORKSPACE removed)

## Verification

Test the migration:

```bash
# Verify Bzlmod is enabled
bazel info | grep bzlmod

# Build and test
bazel build //...
bazel test //...

# Generate dependency graph
bazel mod graph
```
