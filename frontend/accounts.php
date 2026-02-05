<?php
$pageTitle = '帳戶管理';
$currentPage = 'accounts';
include __DIR__ . '/components/header.php';
?>

<ul id="account-list" class="account-list"></ul>

<div id="total-bar" class="total-bar">
    <span>總資產</span>
    <span id="total-balance">$0</span>
</div>

<button class="btn btn-primary" id="btn-add-account">+ 新增帳戶</button>

<!-- 新增/編輯帳戶彈窗 -->
<div class="modal-overlay" id="modal">
    <div class="modal-box">
        <h3 id="modal-title">新增帳戶</h3>
        <input type="text" id="modal-name" placeholder="帳戶名稱">
        <input type="number" id="modal-balance" placeholder="初始餘額" step="1">
        <div class="modal-actions">
            <button class="btn-cancel" id="modal-cancel">取消</button>
            <button class="btn-confirm" id="modal-confirm">確定</button>
        </div>
    </div>
</div>

<script>
    let editingId = null;

    // 載入帳戶列表
    async function loadAccounts() {
        try {
            const accounts = await API.getAccounts();
            const listEl = document.getElementById('account-list');

            if (!accounts || accounts.length === 0) {
                listEl.innerHTML = '<li class="empty-message">尚無帳戶</li>';
                return;
            }

            let totalBalance = 0;
            listEl.innerHTML = accounts.map(a => {
                totalBalance += a.balance;
                const balanceClass = a.balance < 0 ? 'negative' : '';
                return `
                    <li class="account-item" data-id="${a.id}" data-name="${escapeAttr(a.name)}" data-balance="${a.balance}">
                        <span class="account-name">${escapeHtml(a.name)}</span>
                        <span class="account-balance ${balanceClass}">$${Number(a.balance).toLocaleString()}</span>
                    </li>
                `;
            }).join('');

            document.getElementById('total-balance').textContent =
                '$' + Number(totalBalance).toLocaleString();

            // 綁定點擊編輯
            listEl.querySelectorAll('.account-item').forEach(el => {
                el.addEventListener('click', () => {
                    editingId = el.dataset.id;
                    document.getElementById('modal-title').textContent = '編輯帳戶';
                    document.getElementById('modal-name').value = el.dataset.name;
                    document.getElementById('modal-balance').value = el.dataset.balance;
                    document.getElementById('modal').classList.add('active');
                });
            });
        } catch (e) {
            showToast('載入帳戶失敗');
        }
    }

    // 新增帳戶按鈕
    document.getElementById('btn-add-account').addEventListener('click', () => {
        editingId = null;
        document.getElementById('modal-title').textContent = '新增帳戶';
        document.getElementById('modal-name').value = '';
        document.getElementById('modal-balance').value = '0';
        document.getElementById('modal').classList.add('active');
    });

    // 彈窗取消
    document.getElementById('modal-cancel').addEventListener('click', () => {
        document.getElementById('modal').classList.remove('active');
    });

    // 彈窗確定
    document.getElementById('modal-confirm').addEventListener('click', async () => {
        const name = document.getElementById('modal-name').value.trim();
        const balance = parseFloat(document.getElementById('modal-balance').value) || 0;

        if (!name) {
            showToast('請輸入帳戶名稱');
            return;
        }

        try {
            if (editingId) {
                await API.updateAccount(editingId, { name, balance });
                showToast('更新成功');
            } else {
                await API.createAccount({ name, balance });
                showToast('新增成功');
            }
            document.getElementById('modal').classList.remove('active');
            loadAccounts();
        } catch (e) {
            showToast(e.message || '操作失敗');
        }
    });

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    function escapeAttr(text) {
        return text.replace(/"/g, '&quot;').replace(/'/g, '&#39;');
    }

    loadAccounts();
</script>

<?php include __DIR__ . '/components/footer.php'; ?>
