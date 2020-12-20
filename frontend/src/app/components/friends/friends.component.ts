import {Component, OnInit} from '@angular/core';

import {User} from '@models/user';
import {UserService} from '@services/user.service';
import {finalize} from 'rxjs/operators';

@Component({templateUrl: './friends.component.html'})
export class FriendsComponent implements OnInit {
    initiated = false;
    users: User[];

    constructor(private userService: UserService) {
    }

    ngOnInit() {
        this.userService.getFriends()
            .pipe(finalize(() => this.initiated = true))
            .subscribe(r => this.users = r.data);
    }
}
