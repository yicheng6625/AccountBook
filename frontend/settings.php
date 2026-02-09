<?php
$pageTitle = '設定';
$currentPage = 'settings';
include __DIR__ . '/components/header.php';
?>

<div style="padding:12px 16px;font-size:13px;color:#999;background:#f9f9f9;border-bottom:1px solid #eee;">
    分類管理
</div>

<div id="category-list"></div>

<button class="btn btn-primary" id="btn-add-category">+ 新增分類</button>

<!-- 新增/編輯分類彈窗 -->
<div class="modal-overlay" id="modal">
    <div class="modal-box">
        <h3 id="modal-title">新增分類</h3>
        <input type="text" id="modal-name" placeholder="分類名稱">
        <div class="modal-actions">
            <button class="btn-cancel" id="modal-cancel">取消</button>
            <button class="btn-confirm" id="modal-confirm">確定</button>
        </div>
    </div>
</div>

<script>
    let editingId = null;

    // 載入分類列表
    async function loadCategories() {
        try {
            const categories = await API.getCategories();
            const listEl = document.getElementById('category-list');

            if (!categories || categories.length === 0) {
                listEl.innerHTML = '<div class="empty-message">尚無分類</div>';
                return;
            }

            listEl.innerHTML = categories.map(c => `
                <div class="setting-item" data-id="${c.id}">
                    <span class="setting-name">${escapeHtml(c.name)}</span>
                    <div class="setting-actions">
                        <button onclick="editCategory(${c.id}, '${escapeAttr(c.name)}')">編輯</button>
                        <button class="btn-del" onclick="deleteCategory(${c.id}, '${escapeAttr(c.name)}')">刪除</button>
                    </div>
                </div>
            `).join('');
        } catch (e) {
            showToast('載入分類失敗');
        }
    }

    // 編輯分類
    function editCategory(id, name) {
        editingId = id;
        document.getElementById('modal-title').textContent = '編輯分類';
        document.getElementById('modal-name').value = name;
        document.getElementById('modal').classList.add('active');
    }

    // 刪除分類
    async function deleteCategory(id, name) {
        if (!confirm(`確定要刪除分類「${name}」嗎？`)) return;
        try {
            await API.deleteCategory(id);
            showToast('刪除成功');
            loadCategories();
        } catch (e) {
            showToast(e.message || '刪除失敗');
        }
    }

    // 新增分類按鈕
    document.getElementById('btn-add-category').addEventListener('click', () => {
        editingId = null;
        document.getElementById('modal-title').textContent = '新增分類';
        document.getElementById('modal-name').value = '';
        document.getElementById('modal').classList.add('active');
    });

    // 彈窗取消
    document.getElementById('modal-cancel').addEventListener('click', () => {
        document.getElementById('modal').classList.remove('active');
    });

    // 彈窗確定
    document.getElementById('modal-confirm').addEventListener('click', async () => {
        const name = document.getElementById('modal-name').value.trim();
        if (!name) {
            showToast('請輸入分類名稱');
            return;
        }

        try {
            if (editingId) {
                await API.updateCategory(editingId, { name });
                showToast('更新成功');
            } else {
                await API.createCategory({ name });
                showToast('新增成功');
            }
            document.getElementById('modal').classList.remove('active');
            loadCategories();
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
        return text.replace(/'/g, "\\'").replace(/"/g, '\\"');
    }

    loadCategories();
</script>

<?php include __DIR__ . '/components/footer.php'; ?>
