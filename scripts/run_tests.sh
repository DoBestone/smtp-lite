#!/bin/bash

# SMTP Lite 测试运行脚本
# 用法: ./run_tests.sh [unit|bench|cover|all]

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "=== SMTP Lite 测试 ==="
echo

run_unit_tests() {
    echo "📦 运行单元测试..."
    go test ./internal/... -v -count=1
    echo "✅ 单元测试完成"
    echo
}

run_benchmarks() {
    echo "⚡ 运行基准测试..."
    go test ./internal/service/... -bench=. -benchmem -run=^$
    echo "✅ 基准测试完成"
    echo
}

run_coverage() {
    echo "📊 运行覆盖率测试..."
    go test ./internal/... -coverprofile=coverage.out -covermode=atomic
    echo "覆盖率报告:"
    go tool cover -func=coverage.out | tail -1
    echo
    echo "生成 HTML 报告: coverage.html"
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ 覆盖率测试完成"
    echo
}

case "${1:-all}" in
    unit)
        run_unit_tests
        ;;
    bench)
        run_benchmarks
        ;;
    cover)
        run_coverage
        ;;
    all)
        run_unit_tests
        run_benchmarks
        run_coverage
        ;;
    *)
        echo "用法: $0 [unit|bench|cover|all]"
        exit 1
        ;;
esac

echo "🎉 所有测试完成！"