import curses
from curses import panel

def longest_in_the_list(list2):
    # return the size of the longest string in the list
    longest=0
    for item in list2:
        if len(item)>longest:
            longest=len(item)
    return longest

def display_box(list1):
    max_x = longest_in_the_list(list1) + 4
    max_y = len(list1)+2
    w1=curses.newwin(max_y,max_x,curses.LINES/2-len(list1)/2,curses.COLS/2-max_x/2)
    w1panel=panel.new_panel(w1)
    w1.box()
    yy, xx=w1.getmaxyx()
    line=1
    for item in list1:
        w1.addstr(line,2,str(item))
        line+=1
    w1.move(0,0)
    w1.addstr("[X]")
    w1.refresh()
    w1.getch()
    del w1panel

def display_menu(ws,x1,y1,menu_items,attribut1):
    current_option=0
    for o in menu_items:
        ws.addstr(y1,x1,str(o),attribut1[current_option])
        ws.clrtoeol()
        y1+=1
        current_option+=1
    ws.move(2,0)
    ws.refresh()

def run_menu(menu_items,x=0,y=0, subMenu=False):
    """
    will display the menu at x, y on a newly created window
    then display menu to relative coordinates in that new window called wmenu
    see display_menu above
    """
    max_length = longest_in_the_list(menu_items)
    max_option = len(menu_items)
    current_option=0
    option_selected=-1
    wmenu=curses.newwin(max_option+2,max_length,y,x)
    wmenu.border()
    menupanel = panel.new_panel(wmenu)
    color=curses.COLOR_RED
    curses.start_color()
    curses.init_pair(color, curses.COLOR_WHITE, curses.COLOR_BLACK)
    wmenu.bkgdset(ord(' '), curses.color_pair(color))
    wmenu.keypad(1)
    wmenu.refresh()
    while option_selected == -1:
        attribut=[curses.A_NORMAL]*max_option
        attribut[current_option]=curses.A_REVERSE+curses.A_BOLD
        display_menu(wmenu,2,1,menu_items,attribut)
        a=wmenu.getch()
        if   a==curses.KEY_DOWN:
            current_option+=1
        elif a==curses.KEY_UP:
            current_option-=1
        elif a==ord('\n') or a == 32 :
        # validation can be done by CR or space bar
            option_selected=current_option
            if subMenu:
                del menupanel
                panel.update_panels()
        elif a in range(ord('0'),ord('0')+max_option):
        # in case key pressed is a number
            current_option=a-ord('0')
            option_selected=current_option
            if subMenu:
                del menupanel
                panel.update_panels()
        if current_option>max_option-1:
            current_option=max_option-1
        elif current_option <0:
            current_option=0
    return option_selected
