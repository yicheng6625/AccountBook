/**
 * 行事曆元件
 * 原因：首頁上半部的月曆，點擊日期顯示當日紀錄
 */
class Calendar {
    constructor(container, onDateSelect) {
        this.container = container;
        this.onDateSelect = onDateSelect;
        this.today = new Date();
        this.currentYear = this.today.getFullYear();
        this.currentMonth = this.today.getMonth();
        this.selectedDate = this.formatDate(this.today);
        this.dailySummary = {};
        this.render();
    }

    // 格式化日期為 YYYY-MM-DD
    formatDate(date) {
        const y = date.getFullYear();
        const m = String(date.getMonth() + 1).padStart(2, '0');
        const d = String(date.getDate()).padStart(2, '0');
        return `${y}-${m}-${d}`;
    }

    // 格式化月份為 YYYY-MM
    formatMonth() {
        const m = String(this.currentMonth + 1).padStart(2, '0');
        return `${this.currentYear}-${m}`;
    }

    // 切換到上個月
    prevMonth() {
        this.currentMonth--;
        if (this.currentMonth < 0) {
            this.currentMonth = 11;
            this.currentYear--;
        }
        this.loadMonthData();
    }

    // 切換到下個月
    nextMonth() {
        this.currentMonth++;
        if (this.currentMonth > 11) {
            this.currentMonth = 0;
            this.currentYear++;
        }
        this.loadMonthData();
    }

    // 載入月份資料
    async loadMonthData() {
        try {
            const data = await API.getRecordsByMonth(this.formatMonth());
            this.dailySummary = data.daily_summary || {};
        } catch (e) {
            this.dailySummary = {};
        }
        this.render();
    }

    // 選擇日期
    selectDate(dateStr) {
        this.selectedDate = dateStr;
        this.render();
        if (this.onDateSelect) {
            this.onDateSelect(dateStr);
        }
    }

    // 渲染行事曆
    render() {
        const weekdays = ['日', '一', '二', '三', '四', '五', '六'];
        const monthLabel = `${this.currentYear}年${this.currentMonth + 1}月`;

        // 計算月份天數
        const firstDay = new Date(this.currentYear, this.currentMonth, 1);
        const lastDay = new Date(this.currentYear, this.currentMonth + 1, 0);
        const startDayOfWeek = firstDay.getDay();
        const daysInMonth = lastDay.getDate();

        // 上個月的尾端天數
        const prevMonthLast = new Date(this.currentYear, this.currentMonth, 0).getDate();

        let html = `
            <div class="calendar-header">
                <button id="cal-prev">◀</button>
                <span class="month-label">${monthLabel}</span>
                <button id="cal-next">▶</button>
            </div>
            <div class="calendar-grid">
        `;

        // 星期標頭
        weekdays.forEach(w => {
            html += `<div class="weekday">${w}</div>`;
        });

        // 上月尾端填充
        for (let i = startDayOfWeek - 1; i >= 0; i--) {
            html += `<div class="day other-month">${prevMonthLast - i}</div>`;
        }

        // 當月日期
        const todayStr = this.formatDate(this.today);
        for (let d = 1; d <= daysInMonth; d++) {
            const dateStr = `${this.currentYear}-${String(this.currentMonth + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
            let classes = 'day';
            if (dateStr === todayStr) classes += ' today';
            if (dateStr === this.selectedDate) classes += ' selected';
            if (this.dailySummary[dateStr]) classes += ' has-record';
            html += `<div class="${classes}" data-date="${dateStr}">${d}</div>`;
        }

        // 下月開頭填充
        const totalCells = startDayOfWeek + daysInMonth;
        const remaining = totalCells % 7 === 0 ? 0 : 7 - (totalCells % 7);
        for (let i = 1; i <= remaining; i++) {
            html += `<div class="day other-month">${i}</div>`;
        }

        html += '</div>';
        this.container.innerHTML = html;

        // 綁定事件
        this.container.querySelector('#cal-prev').addEventListener('click', () => this.prevMonth());
        this.container.querySelector('#cal-next').addEventListener('click', () => this.nextMonth());

        this.container.querySelectorAll('.day:not(.other-month)').forEach(el => {
            el.addEventListener('click', () => {
                const dateStr = el.dataset.date;
                if (dateStr) this.selectDate(dateStr);
            });
        });
    }
}
