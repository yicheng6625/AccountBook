/**
 * API 呼叫封裝
 * 原因：統一管理所有與後端的通訊邏輯
 */
const API = {
    baseURL: '/api',

    // 通用 fetch 封裝
    async request(endpoint, options = {}) {
        const url = this.baseURL + endpoint;
        const config = {
            headers: { 'Content-Type': 'application/json' },
            ...options,
        };

        if (config.body && typeof config.body === 'object') {
            config.body = JSON.stringify(config.body);
        }

        const response = await fetch(url, config);
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || '請求失敗');
        }
        return data;
    },

    // ========== 紀錄 ==========

    // 查詢指定日期的紀錄
    getRecordsByDate(date) {
        return this.request(`/records?date=${date}`);
    },

    // 查詢指定月份的每日摘要（行事曆用）
    getRecordsByMonth(month) {
        return this.request(`/records?month=${month}`);
    },

    // 取得單筆紀錄
    getRecord(id) {
        return this.request(`/records/${id}`);
    },

    // 新增紀錄
    createRecord(data) {
        return this.request('/records', { method: 'POST', body: data });
    },

    // 更新紀錄
    updateRecord(id, data) {
        return this.request(`/records/${id}`, { method: 'PUT', body: data });
    },

    // 刪除紀錄
    deleteRecord(id) {
        return this.request(`/records/${id}`, { method: 'DELETE' });
    },

    // ========== 帳戶 ==========

    getAccounts() {
        return this.request('/accounts');
    },

    getAccount(id) {
        return this.request(`/accounts/${id}`);
    },

    createAccount(data) {
        return this.request('/accounts', { method: 'POST', body: data });
    },

    updateAccount(id, data) {
        return this.request(`/accounts/${id}`, { method: 'PUT', body: data });
    },

    deleteAccount(id) {
        return this.request(`/accounts/${id}`, { method: 'DELETE' });
    },

    // ========== 分類 ==========

    getCategories() {
        return this.request('/categories');
    },

    createCategory(data) {
        return this.request('/categories', { method: 'POST', body: data });
    },

    updateCategory(id, data) {
        return this.request(`/categories/${id}`, { method: 'PUT', body: data });
    },

    deleteCategory(id) {
        return this.request(`/categories/${id}`, { method: 'DELETE' });
    },

    // ========== 統計 ==========

    getStatistics(month, { accountId, categoryId } = {}) {
        let url = `/statistics?month=${month}`;
        if (accountId) url += `&account_id=${accountId}`;
        if (categoryId) url += `&category_id=${categoryId}`;
        return this.request(url);
    },

    getSummary(month) {
        return this.request(`/statistics/summary?month=${month}`);
    },
};

// 提示訊息工具
function showToast(message) {
    let toast = document.querySelector('.toast');
    if (!toast) {
        toast = document.createElement('div');
        toast.className = 'toast';
        document.body.appendChild(toast);
    }
    toast.textContent = message;
    toast.classList.remove('show');
    // 觸發重排以重新播放動畫
    void toast.offsetWidth;
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 2000);
}

// 格式化金額
function formatAmount(amount, type) {
    const prefix = type === '支出' ? '-$' : '+$';
    return prefix + Number(amount).toLocaleString();
}
