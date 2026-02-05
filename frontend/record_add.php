<?php
$pageTitle = '新增紀錄';
$currentPage = 'add';
include __DIR__ . '/components/header.php';
?>

<form id="add-form">
    <div class="form-group">
        <label>日期</label>
        <input type="date" id="record-date" required>
        <div class="date-shortcuts">
            <button type="button" data-offset="-1">昨天</button>
            <button type="button" data-offset="0" class="active">今天</button>
            <button type="button" data-offset="1">明天</button>
        </div>
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
        <input type="number" id="record-amount" step="1" min="0" required placeholder="請輸入金額">
    </div>
    <div class="form-group">
        <label>項目名稱</label>
        <input type="text" id="record-item" required placeholder="例：午餐">
    </div>
    <div class="form-group">
        <label>分類</label>
        <select id="record-category"></select>
    </div>
    <div class="form-group">
        <label>備註</label>
        <textarea id="record-note" placeholder="選填"></textarea>
    </div>
</form>

<button class="btn btn-primary" id="btn-add">新增</button>

<script>
    // 設定預設日期為今天
    function setDateByOffset(offset) {
        const d = new Date();
        d.setDate(d.getDate() + offset);
        const dateStr = d.toISOString().split('T')[0];
        document.getElementById('record-date').value = dateStr;

        // 更新快捷按鈕狀態
        document.querySelectorAll('.date-shortcuts button').forEach(btn => {
            btn.classList.toggle('active', parseInt(btn.dataset.offset) === offset);
        });
    }

    setDateByOffset(0);

    // 日期快捷按鈕
    document.querySelectorAll('.date-shortcuts button').forEach(btn => {
        btn.addEventListener('click', () => {
            setDateByOffset(parseInt(btn.dataset.offset));
        });
    });

    // 載入帳戶與分類選單
    async function init() {
        try {
            const [accounts, categories] = await Promise.all([
                API.getAccounts(),
                API.getCategories(),
            ]);

            const accountSelect = document.getElementById('record-account');
            (accounts || []).forEach(a => {
                const opt = document.createElement('option');
                opt.value = a.id;
                opt.textContent = a.name;
                accountSelect.appendChild(opt);
            });

            const categorySelect = document.getElementById('record-category');
            (categories || []).forEach(c => {
                const opt = document.createElement('option');
                opt.value = c.id;
                opt.textContent = c.name;
                categorySelect.appendChild(opt);
            });
        } catch (e) {
            showToast('載入資料失敗');
        }
    }

    // 新增紀錄
    document.getElementById('btn-add').addEventListener('click', async () => {
        const data = {
            date: document.getElementById('record-date').value,
            account_id: parseInt(document.getElementById('record-account').value),
            type: document.querySelector('input[name="type"]:checked').value,
            amount: parseFloat(document.getElementById('record-amount').value),
            item: document.getElementById('record-item').value,
            category_id: parseInt(document.getElementById('record-category').value),
            note: document.getElementById('record-note').value,
        };

        if (!data.date || !data.amount || !data.item) {
            showToast('請填寫必要欄位');
            return;
        }

        try {
            await API.createRecord(data);
            showToast('新增成功');
            setTimeout(() => window.location.href = '/', 500);
        } catch (e) {
            showToast(e.message || '新增失敗');
        }
    });

    init();
</script>

<?php include __DIR__ . '/components/footer.php'; ?>
