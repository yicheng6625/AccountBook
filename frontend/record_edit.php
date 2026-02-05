<?php
$pageTitle = '編輯紀錄';
$currentPage = '';
$showBack = true;
$backUrl = '/';
include __DIR__ . '/components/header.php';
?>

<form id="edit-form">
    <div class="form-group">
        <label>日期</label>
        <input type="date" id="record-date" required>
    </div>
    <div class="form-group">
        <label>帳戶</label>
        <select id="record-account"></select>
    </div>
    <div class="form-group">
        <label>類型</label>
        <div class="type-toggle">
            <label><input type="radio" name="type" value="支出" checked> 支出</label>
            <label><input type="radio" name="type" value="收入"> 收入</label>
        </div>
    </div>
    <div class="form-group">
        <label>金額</label>
        <input type="number" id="record-amount" step="1" min="0" required>
    </div>
    <div class="form-group">
        <label>項目名稱</label>
        <input type="text" id="record-item" required>
    </div>
    <div class="form-group">
        <label>分類</label>
        <select id="record-category"></select>
    </div>
    <div class="form-group">
        <label>備註</label>
        <textarea id="record-note"></textarea>
    </div>
</form>

<div class="btn-actions">
    <button class="btn btn-primary" id="btn-save">儲存</button>
    <button class="btn btn-danger" id="btn-delete">刪除</button>
</div>

<script>
    const urlParams = new URLSearchParams(window.location.search);
    const recordId = urlParams.get('id');

    // 載入帳戶與分類選單，以及紀錄資料
    async function init() {
        try {
            const [accounts, categories, record] = await Promise.all([
                API.getAccounts(),
                API.getCategories(),
                API.getRecord(recordId),
            ]);

            // 填入帳戶選單
            const accountSelect = document.getElementById('record-account');
            (accounts || []).forEach(a => {
                const opt = document.createElement('option');
                opt.value = a.id;
                opt.textContent = a.name;
                accountSelect.appendChild(opt);
            });

            // 填入分類選單
            const categorySelect = document.getElementById('record-category');
            (categories || []).forEach(c => {
                const opt = document.createElement('option');
                opt.value = c.id;
                opt.textContent = c.name;
                categorySelect.appendChild(opt);
            });

            // 填入紀錄資料
            document.getElementById('record-date').value = record.date;
            accountSelect.value = record.account_id;
            document.querySelector(`input[name="type"][value="${record.type}"]`).checked = true;
            document.getElementById('record-amount').value = record.amount;
            document.getElementById('record-item').value = record.item;
            categorySelect.value = record.category_id;
            document.getElementById('record-note').value = record.note || '';
        } catch (e) {
            showToast('載入紀錄失敗');
        }
    }

    // 儲存
    document.getElementById('btn-save').addEventListener('click', async () => {
        const data = {
            date: document.getElementById('record-date').value,
            account_id: parseInt(document.getElementById('record-account').value),
            type: document.querySelector('input[name="type"]:checked').value,
            amount: parseFloat(document.getElementById('record-amount').value),
            item: document.getElementById('record-item').value,
            category_id: parseInt(document.getElementById('record-category').value),
            note: document.getElementById('record-note').value,
        };

        try {
            await API.updateRecord(recordId, data);
            showToast('更新成功');
            setTimeout(() => window.location.href = '/', 500);
        } catch (e) {
            showToast(e.message || '更新失敗');
        }
    });

    // 刪除
    document.getElementById('btn-delete').addEventListener('click', async () => {
        if (!confirm('確定要刪除這筆紀錄嗎？')) return;

        try {
            await API.deleteRecord(recordId);
            showToast('刪除成功');
            setTimeout(() => window.location.href = '/', 500);
        } catch (e) {
            showToast(e.message || '刪除失敗');
        }
    });

    init();
</script>

<?php include __DIR__ . '/components/footer.php'; ?>
