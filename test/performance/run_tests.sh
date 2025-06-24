#!/bin/bash

# 设置环境变量
export BASE_URL=${BASE_URL:-"http://localhost:8080"}
export K6_PROMETHEUS_RW_SERVER_URL=${K6_PROMETHEUS_RW_SERVER_URL:-"http://localhost:9090/api/v1/write"}

# 颜色输出函数
print_green() {
    echo -e "\033[0;32m$1\033[0m"
}

print_yellow() {
    echo -e "\033[1;33m$1\033[0m"
}

print_red() {
    echo -e "\033[0;31m$1\033[0m"
}

# 运行测试函数
run_test() {
    local test_name=$1
    local script_path=$2
    local extra_args=${3:-""}

    print_yellow "开始运行 $test_name..."
    print_yellow "使用API地址: $BASE_URL"
    
    # 运行测试
    if [ -n "$K6_PROMETHEUS_RW_SERVER_URL" ]; then
        # 如果设置了 Prometheus 地址，同时输出到 Prometheus 和控制台
        k6 run \
            --out experimental-prometheus-rw \
            --out json=test_results/${test_name}.json \
            --tag testname=$test_name \
            $extra_args \
            scenarios/$script_path
    else
        # 否则只输出到控制台和JSON文件
        k6 run \
            --out json=test_results/${test_name}.json \
            --tag testname=$test_name \
            $extra_args \
            scenarios/$script_path
    fi

    if [ $? -eq 0 ]; then
        print_green "$test_name 测试完成"
        print_green "测试结果已保存到 test_results/${test_name}.json"
    else
        print_red "$test_name 测试失败"
        exit 1
    fi

    # 测试之间休息一下，让系统恢复
    sleep 30
}

# 主函数
main() {
    # 检查k6是否安装
    if ! command -v k6 &> /dev/null; then
        print_red "错误: k6 未安装。请先安装 k6: https://k6.io/docs/getting-started/installation"
        exit 1
    fi

    # 创建结果目录
    mkdir -p test_results

    print_yellow "开始性能测试..."

    # 场景一：只读接口基准测试
    run_test "read-only-test" "read_only_test.js"

    # 场景二：混合读写测试
    run_test "real-user-simulation" "real_user_simulation.js"

    # 场景三：高并发抢票测试
    # 注意：这里需要确保TEST_SHOWTIME_ID存在且有足够的座位
    run_test "rush-booking-test" "rush_booking_test.js" "--env SHOWTIME_ID=101"

    # 场景四：长时间稳定性测试
    # 注意：这个测试会运行8小时
    print_yellow "准备开始长时间稳定性测试（8小时）..."
    read -p "是否继续？[y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        run_test "soak-test" "soak_test.js"
    fi

    print_green "所有测试完成！"
    print_yellow "测试结果保存在 test_results 目录中"
}

# 运行主函数
main 