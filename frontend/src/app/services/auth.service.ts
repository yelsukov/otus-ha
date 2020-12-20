import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {BehaviorSubject, Observable} from 'rxjs';
import {map} from 'rxjs/operators';

import {environment} from '@env/environment';
import {Session, User} from '@models/user';

@Injectable({providedIn: 'root'})
export class AuthService {
    private sessionSubject: BehaviorSubject<Session>;
    public currentSession: Observable<Session>;

    constructor(private http: HttpClient) {
        this.sessionSubject = new BehaviorSubject<Session>(JSON.parse(localStorage.getItem('currentSession')));
        this.currentSession = this.sessionSubject.asObservable();
    }

    public get currentSessionValue(): Session {
        return this.sessionSubject.value;
    }

    login(username: string, password: string) {
        return this.http.post<any>(`${environment.apiUrl}/auth/sign-in`, {username, password})
            .pipe(map(session => {
                localStorage.setItem('currentSession', JSON.stringify(session));
                this.sessionSubject.next(session);
                return session;
            }));
    }

    signUp(u: User) {
        if (!!u.age && typeof u.age === 'string') {
            delete u.age;
        }
        return this.http.post(environment.apiUrl + '/auth/sign-up', u);
    }

    logout() {
        localStorage.removeItem('currentSession');
        this.sessionSubject.next(null);
    }
}
