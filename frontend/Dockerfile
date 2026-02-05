# 使用 Debian 版本，curl 擴充已內建
# 原因：避免 Alpine DNS 解析問題
FROM php:8.3-cli

# 複製 PHP 程式碼
COPY . /var/www/html

WORKDIR /var/www/html

EXPOSE 80

# 使用 PHP 內建伺服器搭配路由器，透過 Cloudflared Tunnel 對外
CMD ["php", "-S", "0.0.0.0:80", "router.php"]
