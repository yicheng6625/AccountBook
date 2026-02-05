<?php
/**
 * PHP 內建伺服器路由器
 * 原因：不使用 Nginx，需自行處理路由分發與 API 代理
 */

$uri = $_SERVER['REQUEST_URI'];
$path = parse_url($uri, PHP_URL_PATH);

// API 代理：將 /api/ 請求轉發至 Go 後端
if (strpos($path, '/api/') === 0) {
    $backendUrl = 'http://backend:8080' . $uri;

    $method = $_SERVER['REQUEST_METHOD'];
    $headers = ['Content-Type: application/json'];

    $ch = curl_init($backendUrl);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
    curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);

    // 若為 POST/PUT，轉發 request body
    if ($method === 'POST' || $method === 'PUT') {
        $input = file_get_contents('php://input');
        curl_setopt($ch, CURLOPT_POSTFIELDS, $input);
    }

    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    http_response_code($httpCode);
    header('Content-Type: application/json');
    echo $response;
    return true;
}

// 靜態檔案（CSS、JS）直接返回
if (preg_match('/\.(css|js|png|jpg|gif|ico|svg)$/', $path)) {
    return false;
}

// PHP 頁面路由
if ($path === '/' || $path === '/index.php') {
    include __DIR__ . '/index.php';
    return true;
}

$file = __DIR__ . $path;
if (file_exists($file) && pathinfo($file, PATHINFO_EXTENSION) === 'php') {
    include $file;
    return true;
}

// 404
http_response_code(404);
echo '頁面不存在';
return true;
