import {Component, OnInit} from '@angular/core';

import {User} from '@models/user';
import {UserService} from '@services/user.service';

import {Event} from '@models/event';
import {NewsService} from '@services/news.service';

@Component({templateUrl: 'dashboard.component.html'})
export class DashboardComponent implements OnInit {
    loading = false;
    loadingNews = false;
    users: User[];
    news: Event[];

    constructor(private userService: UserService, private newsService: NewsService) {
    }

    ngOnInit() {
        this.loading = true;
        this.userService.getAll().subscribe(r => {
            this.loading = false;
            this.users = r.data;
        });
        this.loadingNews = true;
        this.newsService.getAll().subscribe(r => {
            this.loadingNews = false;
            this.news = r.data;
        });
    }
}
