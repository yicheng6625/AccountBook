<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title><?= $pageTitle ?? '記帳本' ?></title>
    <link rel="stylesheet" href="/assets/css/style.css">
</head>
<body>
<div class="app-container">
    <header class="app-header">
        <?php if (isset($showBack) && $showBack): ?>
            <a href="<?= $backUrl ?? '/' ?>" class="back-btn">←</a>
        <?php endif; ?>
        <h1><?= $pageTitle ?? '記帳本' ?></h1>
    </header>
    <main class="app-main">
