#!/bin/bash

echo "测试MCP协议服务器..."

# 启动MCP服务器在后台
echo "启动MCP协议服务器..."
echo '{"jsonrpc": "2.0", "id": 1, "method": "list_tools", "params": {}}' | ./scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080 mcp

echo "MCP测试完成" 