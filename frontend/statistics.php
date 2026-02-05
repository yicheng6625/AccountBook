<?php
$pageTitle = '統計';
$currentPage = 'statistics';
$extraScripts = ['/assets/js/chart.js'];
include __DIR__ . '/components/header.php';
?>

<!-- 引入 Chart.js CDN -->
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"></script>

<div class="calendar-header">
    <button id="stat-prev">◀</button>
    <span class="month-label" id="stat-month-label"></span>
    <button id="stat-next">▶</button>
</div>

<div class="stat-summary">
    <div class="stat-row">
        <div>
            <div class="stat-label">收入</div>
            <div class="stat-value income" id="stat-income">$0</div>
        </div>
        <div>
            <div class="stat-label">支出</div>
            <div class="stat-value expense" id="stat-expense">$0</div>
        </div>
        <div>
            <div class="stat-label">結餘</div>
            <div class="stat-value balance" id="stat-balance">$0</div>
        </div>
    </div>
</div>

<div class="chart-container">
    <canvas id="expense-chart" width="260" height="260"></canvas>
</div>

<div id="category-list" class="category-list"></div>

<script>
    let currentYear = new Date().getFullYear();
    let currentMonth = new Date().getMonth() + 1;

    function formatMonth() {
        return `${currentYear}-${String(currentMonth).padStart(2, '0')}`;
    }

    function updateLabel() {
        document.getElementById('stat-month-label').textContent =
            `${currentYear}年${currentMonth}月`;
    }

    document.getElementById('stat-prev').addEventListener('click', () => {
        currentMonth--;
        if (currentMonth < 1) { currentMonth = 12; currentYear--; }
        loadStatistics();
    });

    document.getElementById('stat-next').addEventListener('click', () => {
        currentMonth++;
        if (currentMonth > 12) { currentMonth = 1; currentYear++; }
        loadStatistics();
    });

    async function loadStatistics() {
        updateLabel();
        const month = formatMonth();

        try {
            const data = await API.getStatistics(month);

            document.getElementById('stat-income').textContent =
                '$' + Number(data.total_income).toLocaleString();
            document.getElementById('stat-expense').textContent =
                '$' + Number(data.total_expense).toLocaleString();
            document.getElementById('stat-balance').textContent =
                '$' + Number(data.total_income - data.total_expense).toLocaleString();

            // 渲染支出圓餅圖
            const categories = data.expense_categories || [];
            renderPieChart('expense-chart', categories, '支出分類');
            renderCategoryList('category-list', categories);
        } catch (e) {
            document.getElementById('stat-income').textContent = '$0';
            document.getElementById('stat-expense').textContent = '$0';
            document.getElementById('stat-balance').textContent = '$0';
            document.getElementById('category-list').innerHTML =
                '<div class="empty-message">無統計資料</div>';
        }
    }

    loadStatistics();
</script>

<?php include __DIR__ . '/components/footer.php'; ?>
