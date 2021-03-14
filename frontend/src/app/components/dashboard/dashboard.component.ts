import {Component, OnDestroy, OnInit} from '@angular/core';

import {User} from '@models/user';
import {UserService} from '@services/user.service';

import {Event} from '@models/event';
import {NewsService} from '@services/news.service';

@Component({templateUrl: 'dashboard.component.html'})
export class DashboardComponent implements OnInit, OnDestroy {
    loading = false;
    loadingNews = false;
    users: User[];
    news: Event[];

    constructor(private userService: UserService, private newsService: NewsService) {
    }

    ngOnInit() {
        const that = this;

        this.loading = true;
        this.userService.getAll().subscribe(r => {
            that.loading = false;
            that.users = r.data;
        });

        if (!this.newsService.isConnected()) {
            that.loadingNews = true;
        }
        this.newsService.connect().subscribe({
            next: (r) => {
                that.loadingNews = false;
                switch (r.object) {
                    case 'list':
                        that.news = r.data;
                        break;
                    case 'event':
                        that.news.unshift(r);
                        break;
                    case 'error':
                        console.log(r.code + ': ' + r.message);
                }
            },
            error: e => console.log(e),
        });
    }

    ngOnDestroy() {
        this.newsService.close();
    }
}
