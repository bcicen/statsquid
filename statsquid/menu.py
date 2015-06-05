import curses
from curses import panel

def run_menu(menu_items,x=0,y=0,name=None):
    """
    will display the menu at x, y on a newly created window
    """
    length = max(len(s) for s in menu_items) + 4
    height = len(menu_items)
    if name:
        height += 2
    current_option = 0
    option_selected = -1

    w = curses.newwin(height+2,length,y,x)
    w.keypad(1)
    w.refresh()

    while option_selected == -1:
        attribut=[curses.A_NORMAL]*height
        attribut[current_option]=curses.A_REVERSE+curses.A_BOLD
        line = 1
        if name:
            w.addstr(line,2,name,curses.A_BOLD+curses.A_UNDERLINE)
            line += 2
        for i,v in enumerate(menu_items):
            w.addstr(line,2,str(v),attribut[i])
            w.clrtoeol()
            line += 1
        w.border()
        w.refresh()
        a = w.getch()
        if a == curses.KEY_DOWN:
            current_option+=1
        elif a == curses.KEY_UP:
            current_option-=1
        elif a==ord('\n') or a == 32 :
        # validation can be done by CR or space bar
            option_selected=current_option
        elif a in range(ord('0'),ord('0')+height):
        # in case key pressed is a number
            current_option=a-ord('0')
            option_selected=current_option
        if current_option>height-1:
            current_option=height-1
        elif current_option <0:
            current_option=0
    return option_selected
