import {Component, OnInit} from '@angular/core';

import { User } from '@models/user';
import { UserService } from '@services/user.service';

@Component({ templateUrl: 'dashboard.component.html' })
export class DashboardComponent implements OnInit {
    loading = false;
    users: User[];

    constructor(private userService: UserService) { }

    ngOnInit() {
        this.loading = true;
        this.userService.getAll().subscribe(r => {
            this.loading = false;
            this.users = r.data;
        });
    }
}
