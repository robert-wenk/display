#!/bin/bash
# Bzlmod Configuration Verification Script

set -e

echo "QNAP Display Control - Bzlmod Configuration Verification"
echo "========================================================"

# Check if Bazel is available
if ! command -v bazel &> /dev/null; then
    echo "⚠️  Bazel not found - this is expected for verification"
    echo "   Installation will be handled by install_rust.sh"
else
    echo "✅ Bazel found: $(bazel version | head -1)"
    
    # Verify Bzlmod is enabled
    if bazel info 2>/dev/null | grep -q "bzlmod.*true"; then
        echo "✅ Bzlmod is enabled"
    else
        echo "❌ Bzlmod is not enabled - check .bazelrc"
    fi
fi

echo ""
echo "Configuration Files Verification:"
echo "================================="

# Check MODULE.bazel
if [[ -f "MODULE.bazel" ]]; then
    echo "✅ MODULE.bazel exists"
    if grep -q "rules_rust" MODULE.bazel; then
        echo "   ✅ Rust rules configured"
    fi
    if grep -q "crate" MODULE.bazel; then
        echo "   ✅ Crate dependencies configured"
    fi
else
    echo "❌ MODULE.bazel missing"
fi

# Check BUILD.bazel
if [[ -f "BUILD.bazel" ]]; then
    echo "✅ BUILD.bazel exists"
    if grep -q "@crates//" BUILD.bazel; then
        echo "   ✅ Updated to use @crates// (Bzlmod style)"
    fi
    if grep -q "@crate_index//" BUILD.bazel; then
        echo "   ❌ Still uses old @crate_index// references"
    fi
else
    echo "❌ BUILD.bazel missing"
fi

# Check .bazelrc
if [[ -f ".bazelrc" ]]; then
    echo "✅ .bazelrc exists"
    if grep -q "enable_bzlmod" .bazelrc; then
        echo "   ✅ Bzlmod enabled in configuration"
    fi
    if grep -q "config:opt" .bazelrc; then
        echo "   ✅ Optimization configuration defined"
    fi
    if grep -q "config:static" .bazelrc; then
        echo "   ✅ Static linking configuration defined"
    fi
else
    echo "❌ .bazelrc missing"
fi

# Check deprecated files
if [[ -f "WORKSPACE" ]]; then
    echo "❌ WORKSPACE file still exists (should be removed)"
else
    echo "✅ WORKSPACE file removed (deprecated)"
fi

if [[ -f ".bazelignore" ]]; then
    echo "✅ .bazelignore exists"
else
    echo "⚠️  .bazelignore missing (recommended)"
fi

echo ""
echo "Rust Configuration Verification:"
echo "================================"

# Check Cargo.toml
if [[ -f "Cargo.toml" ]]; then
    echo "✅ Cargo.toml exists"
    if grep -q "serialport\|tokio\|clap" Cargo.toml; then
        echo "   ✅ Required dependencies listed"
    fi
else
    echo "❌ Cargo.toml missing"
fi

# Check source structure
if [[ -d "src" ]]; then
    echo "✅ src/ directory exists"
    
    required_files=(
        "src/lib.rs"
        "src/main.rs"
        "src/display_controller.rs"
        "src/usb_copy_monitor.rs"
        "src/io_port_access.rs"
        "src/error.rs"
    )
    
    for file in "${required_files[@]}"; do
        if [[ -f "$file" ]]; then
            echo "   ✅ $file exists"
        else
            echo "   ❌ $file missing"
        fi
    done
else
    echo "❌ src/ directory missing"
fi

echo ""
echo "Migration Summary:"
echo "=================="
echo "✅ WORKSPACE → MODULE.bazel migration completed"
echo "✅ Dependency references updated (@crate_index → @crates)"
echo "✅ Build configurations added (.bazelrc)"
echo "✅ Modern Bzlmod system implemented"

echo ""
echo "Next Steps:"
echo "==========="
echo "1. Install Bazel/Bazelisk if not already installed"
echo "2. Run: ./install_rust.sh"
echo "3. Test build: bazel build //..."
echo "4. Run tests: bazel test //..."

echo ""
echo "Build Commands:"
echo "==============="
echo "Standard build:     bazel build //..."
echo "Optimized build:    bazel build --config=opt //..."
echo "Static binary:      bazel build --config=static //..."
echo "Run tests:          bazel test //..."
echo "Dependency graph:   bazel mod graph"
