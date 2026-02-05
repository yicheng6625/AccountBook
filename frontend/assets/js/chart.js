/**
 * 圖表工具
 * 原因：統計頁圓餅圖渲染，使用 Chart.js
 */

// 預設色彩組
const CHART_COLORS = [
    '#4A90D9', '#E74C3C', '#27AE60', '#F39C12',
    '#9B59B6', '#1ABC9C', '#E67E22', '#3498DB',
    '#2ECC71', '#E91E63', '#00BCD4', '#FF9800',
];

/**
 * 渲染圓餅圖
 * @param {string} canvasId - canvas 元素 ID
 * @param {Array} categories - [{name, amount, percentage}]
 * @param {string} title - 圖表標題
 */
function renderPieChart(canvasId, categories, title) {
    const canvas = document.getElementById(canvasId);
    if (!canvas || !categories || categories.length === 0) return;

    const ctx = canvas.getContext('2d');
    const data = categories.map(c => c.amount);
    const labels = categories.map(c => c.name);
    const colors = categories.map((_, i) => CHART_COLORS[i % CHART_COLORS.length]);

    // 若已有圖表實例則銷毀
    if (window._pieChart) {
        window._pieChart.destroy();
    }

    window._pieChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: labels,
            datasets: [{
                data: data,
                backgroundColor: colors,
                borderWidth: 1,
                borderColor: '#fff',
            }],
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false,
                },
                title: {
                    display: !!title,
                    text: title || '',
                    font: { size: 14 },
                },
            },
        },
    });
}

/**
 * 渲染分類統計列表
 * @param {string} containerId - 容器元素 ID
 * @param {Array} categories - [{name, amount, percentage}]
 */
function renderCategoryList(containerId, categories) {
    const container = document.getElementById(containerId);
    if (!container) return;

    if (!categories || categories.length === 0) {
        container.innerHTML = '<div class="empty-message">無資料</div>';
        return;
    }

    container.innerHTML = categories.map((c, i) => {
        const color = CHART_COLORS[i % CHART_COLORS.length];
        return `
            <div class="category-stat-item">
                <span class="color-dot" style="background:${color}"></span>
                <span class="cat-name">${escapeHtml(c.name)}</span>
                <span class="cat-amount">$${Number(c.amount).toLocaleString()}</span>
                <span class="cat-percent">${c.percentage.toFixed(1)}%</span>
            </div>
        `;
    }).join('');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
