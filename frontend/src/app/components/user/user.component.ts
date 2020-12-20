import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {finalize, first} from 'rxjs/operators';

import {AuthService} from '@services/auth.service';
import {UserService} from '@services/user.service';

import {User} from '@models/user';


@Component({templateUrl: './user.component.html'})
export class UserComponent implements OnInit {
    initiated = false;
    loading = false;
    user: User;

    private isCurrentUser = false;

    constructor(private route: ActivatedRoute, private userService: UserService, private auth: AuthService) {
    }

    ngOnInit(): void {
        const userId = parseInt(this.route.snapshot.paramMap.get('id'), 10);
        this.isCurrentUser = userId === this.auth.currentSessionValue.userId;
        this.userService.get(userId)
            .pipe(finalize(() => this.initiated = true))
            .subscribe(user => this.user = user);
    }

    makeFriendship(): void {
        this.loading = true;
        this.userService.createFriendship(this.user.id)
            .pipe(first(), finalize(() => this.loading = false))
            .subscribe(() => this.user.isFriend = true);
    }

    breakFriendship(): void {
        this.loading = true;
        this.userService.breakFriendship(this.user.id)
            .pipe(first(), finalize(() => this.loading = false))
            .subscribe(() => this.user.isFriend = false);
    }
}
