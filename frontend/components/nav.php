    </main>
    <nav class="bottom-nav">
        <a href="/" class="nav-item <?= ($currentPage ?? '') === 'home' ? 'active' : '' ?>">
            <span class="nav-icon">📅</span>
            <span class="nav-label">首頁</span>
        </a>
        <a href="/record_add.php" class="nav-item <?= ($currentPage ?? '') === 'add' ? 'active' : '' ?>">
            <span class="nav-icon">＋</span>
            <span class="nav-label">新增</span>
        </a>
        <a href="/accounts.php" class="nav-item <?= ($currentPage ?? '') === 'accounts' ? 'active' : '' ?>">
            <span class="nav-icon">💰</span>
            <span class="nav-label">帳戶</span>
        </a>
        <a href="/statistics.php" class="nav-item <?= ($currentPage ?? '') === 'statistics' ? 'active' : '' ?>">
            <span class="nav-icon">📊</span>
            <span class="nav-label">統計</span>
        </a>
        <a href="/settings.php" class="nav-item <?= ($currentPage ?? '') === 'settings' ? 'active' : '' ?>">
            <span class="nav-icon">⚙</span>
            <span class="nav-label">設定</span>
        </a>
    </nav>
</div>
